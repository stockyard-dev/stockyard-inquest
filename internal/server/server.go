package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/stockyard-dev/stockyard-inquest/internal/store"
)

type Server struct {
	db     *store.DB
	mux    *http.ServeMux
	limits Limits
}

func New(db *store.DB, limits Limits) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), limits: limits}

	s.mux.HandleFunc("GET /api/incidents", s.listIncidents)
	s.mux.HandleFunc("POST /api/incidents", s.createIncident)
	s.mux.HandleFunc("GET /api/incidents/{id}", s.getIncident)
	s.mux.HandleFunc("PUT /api/incidents/{id}", s.updateIncident)
	s.mux.HandleFunc("DELETE /api/incidents/{id}", s.deleteIncident)
	s.mux.HandleFunc("POST /api/incidents/{id}/status", s.updateStatus)

	s.mux.HandleFunc("GET /api/incidents/{id}/timeline", s.listTimeline)
	s.mux.HandleFunc("POST /api/incidents/{id}/timeline", s.addUpdate)

	s.mux.HandleFunc("GET /api/incidents/{id}/postmortem", s.getPostmortem)
	s.mux.HandleFunc("POST /api/incidents/{id}/postmortem", s.createPostmortem)
	s.mux.HandleFunc("PUT /api/incidents/{id}/postmortem", s.updatePostmortem)

	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/health", s.health)

	s.mux.HandleFunc("GET /ui", s.dashboard)
	s.mux.HandleFunc("GET /ui/", s.dashboard)
	s.mux.HandleFunc("GET /", s.root)
s.mux.HandleFunc("GET /api/tier",func(w http.ResponseWriter,r *http.Request){wj(w,200,map[string]any{"tier":s.limits.Tier,"upgrade_url":"https://stockyard.dev/inquest/"})})

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json"); w.WriteHeader(code); json.NewEncoder(w).Encode(v)
}
func writeErr(w http.ResponseWriter, code int, msg string) { writeJSON(w, code, map[string]string{"error": msg}) }
func (s *Server) root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" { http.NotFound(w, r); return }
	http.Redirect(w, r, "/ui", http.StatusFound)
}

func (s *Server) listIncidents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	f := store.IncidentFilter{Status: q.Get("status"), Severity: q.Get("severity"), Lead: q.Get("lead"), Search: q.Get("search"), Limit: limit, Offset: offset}
	if f.Status == "" { f.Status = "all" }
	incs, total := s.db.ListIncidents(f)
	writeJSON(w, 200, map[string]any{"incidents": orEmpty(incs), "total": total})
}

func (s *Server) createIncident(w http.ResponseWriter, r *http.Request) {
	var inc store.Incident
	if err := json.NewDecoder(r.Body).Decode(&inc); err != nil { writeErr(w, 400, "invalid json"); return }
	if inc.Title == "" { writeErr(w, 400, "title required"); return }
	if err := s.db.CreateIncident(&inc); err != nil { writeErr(w, 500, err.Error()); return }
	writeJSON(w, 201, s.db.GetIncident(inc.ID))
}

func (s *Server) getIncident(w http.ResponseWriter, r *http.Request) {
	inc := s.db.GetIncident(r.PathValue("id"))
	if inc == nil { writeErr(w, 404, "not found"); return }
	writeJSON(w, 200, inc)
}

func (s *Server) updateIncident(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ex := s.db.GetIncident(id)
	if ex == nil { writeErr(w, 404, "not found"); return }
	var inc store.Incident
	if err := json.NewDecoder(r.Body).Decode(&inc); err != nil { writeErr(w, 400, "invalid json"); return }
	if inc.Title == "" { inc.Title = ex.Title }
	if inc.Severity == "" { inc.Severity = ex.Severity }
	if inc.Status == "" { inc.Status = ex.Status }
	if inc.Services == nil { inc.Services = ex.Services }
	if err := s.db.UpdateIncident(id, &inc); err != nil { writeErr(w, 500, err.Error()); return }
	writeJSON(w, 200, s.db.GetIncident(id))
}

func (s *Server) deleteIncident(w http.ResponseWriter, r *http.Request) {
	if err := s.db.DeleteIncident(r.PathValue("id")); err != nil { writeErr(w, 500, err.Error()); return }
	writeJSON(w, 200, map[string]string{"deleted": "ok"})
}

func (s *Server) updateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct{ Status string `json:"status"`; Author string `json:"author"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { writeErr(w, 400, "invalid json"); return }
	if req.Status == "" { writeErr(w, 400, "status required"); return }
	if err := s.db.UpdateStatus(id, req.Status, req.Author); err != nil { writeErr(w, 500, err.Error()); return }
	writeJSON(w, 200, s.db.GetIncident(id))
}

func (s *Server) listTimeline(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"timeline": orEmpty(s.db.ListTimeline(r.PathValue("id")))})
}

func (s *Server) addUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct{ Author string `json:"author"`; Message string `json:"message"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { writeErr(w, 400, "invalid json"); return }
	if req.Message == "" { writeErr(w, 400, "message required"); return }
	if err := s.db.AddUpdate(id, req.Author, req.Message); err != nil { writeErr(w, 500, err.Error()); return }
	writeJSON(w, 201, map[string]string{"added": "ok"})
}

func (s *Server) getPostmortem(w http.ResponseWriter, r *http.Request) {
	pm := s.db.GetPostmortem(r.PathValue("id"))
	if pm == nil { writeJSON(w, 200, map[string]any{"postmortem": nil}); return }
	writeJSON(w, 200, pm)
}

func (s *Server) createPostmortem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if s.db.GetIncident(id) == nil { writeErr(w, 404, "incident not found"); return }
	if s.db.GetPostmortem(id) != nil { writeErr(w, 409, "postmortem already exists, use PUT to update"); return }
	var pm store.Postmortem
	if err := json.NewDecoder(r.Body).Decode(&pm); err != nil { writeErr(w, 400, "invalid json"); return }
	pm.IncidentID = id
	if err := s.db.CreatePostmortem(&pm); err != nil { writeErr(w, 500, err.Error()); return }
	writeJSON(w, 201, pm)
}

func (s *Server) updatePostmortem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if s.db.GetPostmortem(id) == nil { writeErr(w, 404, "postmortem not found"); return }
	var pm store.Postmortem
	if err := json.NewDecoder(r.Body).Decode(&pm); err != nil { writeErr(w, 400, "invalid json"); return }
	if err := s.db.UpdatePostmortem(id, &pm); err != nil { writeErr(w, 500, err.Error()); return }
	writeJSON(w, 200, s.db.GetPostmortem(id))
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, s.db.Stats()) }
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	st := s.db.Stats()
	writeJSON(w, 200, map[string]any{"status": "ok", "service": "inquest", "incidents": st.Total, "active": st.Active})
}
func orEmpty[T any](s []T) []T { if s == nil { return []T{} }; return s }
func init() { log.SetFlags(log.LstdFlags | log.Lshortfile) }
