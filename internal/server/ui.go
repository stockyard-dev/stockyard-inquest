package server

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashHTML))
}

const dashHTML = `<!DOCTYPE html>
<html lang="en"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Inquest</title>
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#c45d2c;--rl:#e8753a;--leather:#a0845c;--ll:#c4a87a;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c44040;--blue:#4a7ec4;--mono:'JetBrains Mono',Consolas,monospace;--serif:'Libre Baskerville',Georgia,serif}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--mono);font-size:13px;line-height:1.6}
a{color:var(--rl);text-decoration:none}a:hover{color:var(--gold)}
.hdr{padding:.6rem 1.2rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}
.hdr h1{font-family:var(--serif);font-size:1rem}.hdr h1 span{color:var(--rl)}
.hdr-right{display:flex;gap:1rem;align-items:center;font-size:.7rem;color:var(--leather)}.hdr-right b{color:var(--cream)}
.main{max-width:900px;margin:0 auto;padding:1rem 1.2rem}
.overview{display:flex;gap:1.5rem;margin-bottom:1rem;font-size:.7rem;color:var(--leather);flex-wrap:wrap}
.overview .stat b{display:block;font-size:1.2rem;color:var(--cream)}
.toolbar{display:flex;gap:.5rem;margin-bottom:.8rem;flex-wrap:wrap;align-items:center}
.toolbar select,.toolbar input{background:var(--bg);border:1px solid var(--bg3);color:var(--cream);padding:.3rem .5rem;font-family:var(--mono);font-size:.72rem;outline:none}
.toolbar select:focus,.toolbar input:focus{border-color:var(--rust)}
.toolbar input{flex:1;min-width:120px}
.btn{font-family:var(--mono);font-size:.68rem;padding:.3rem .6rem;border:1px solid;cursor:pointer;background:transparent;transition:.15s;white-space:nowrap}
.btn-p{border-color:var(--rust);color:var(--rl)}.btn-p:hover{background:var(--rust);color:var(--cream)}
.btn-d{border-color:var(--bg3);color:var(--cm)}.btn-d:hover{border-color:var(--red);color:var(--red)}
.btn-s{border-color:var(--green);color:var(--green)}.btn-s:hover{background:var(--green);color:var(--bg)}
.btn-b{border-color:var(--blue);color:var(--blue)}.btn-b:hover{background:var(--blue);color:var(--cream)}
.btn-g{border-color:var(--gold);color:var(--gold)}.btn-g:hover{background:var(--gold);color:var(--bg)}

.inc-card{background:var(--bg2);border:1px solid var(--bg3);padding:.7rem;margin-bottom:.4rem;cursor:pointer;transition:background .1s}
.inc-card:hover{background:var(--bg3)}
.inc-top{display:flex;align-items:center;gap:.5rem}
.inc-title{font-size:.8rem;font-weight:600;flex:1}
.sev-badge{font-size:.6rem;padding:.1rem .35rem;border:1px solid;border-radius:2px;text-transform:uppercase;font-weight:600}
.sev-sev1{border-color:var(--red);color:var(--red);background:rgba(196,64,64,.1)}.sev-sev2{border-color:var(--rl);color:var(--rl)}.sev-sev3{border-color:var(--gold);color:var(--gold)}.sev-sev4{border-color:var(--cm);color:var(--cm)}
.status-badge{font-size:.6rem;padding:.1rem .35rem;border-radius:2px;text-transform:uppercase}
.st-investigating{background:rgba(196,64,64,.15);color:var(--red)}.st-identified{background:rgba(232,117,58,.15);color:var(--rl)}.st-monitoring{background:rgba(212,168,67,.15);color:var(--gold)}.st-resolved{background:rgba(74,158,92,.15);color:var(--green)}
.inc-meta{font-size:.65rem;color:var(--cm);margin-top:.3rem;display:flex;gap:.8rem;flex-wrap:wrap}

.modal-bg{position:fixed;top:0;left:0;right:0;bottom:0;background:rgba(0,0,0,.65);display:flex;align-items:center;justify-content:center;z-index:100}
.modal{background:var(--bg2);border:1px solid var(--bg3);padding:1.5rem;width:95%;max-width:750px;max-height:90vh;overflow-y:auto}
.modal h2{font-family:var(--serif);font-size:.95rem;margin-bottom:1rem}
label.fl{display:block;font-size:.65rem;color:var(--leather);text-transform:uppercase;letter-spacing:1px;margin-bottom:.25rem;margin-top:.7rem}
input[type=text],textarea,select{background:var(--bg);border:1px solid var(--bg3);color:var(--cream);padding:.4rem .6rem;font-family:var(--mono);font-size:.8rem;width:100%;outline:none}
input:focus,textarea:focus,select:focus{border-color:var(--rust)}
textarea{resize:vertical;min-height:60px}
.form-row{display:flex;gap:.5rem}.form-row>*{flex:1}

.tl-item{display:flex;gap:.6rem;padding:.4rem 0;border-bottom:1px solid var(--bg3);font-size:.72rem}
.tl-dot{width:10px;height:10px;border-radius:50%;background:var(--rust);margin-top:4px;flex-shrink:0}
.tl-content{flex:1}.tl-time{color:var(--cm);font-size:.6rem;margin-left:.5rem}
.tl-status{font-size:.6rem;padding:0 .25rem;border-radius:2px;margin-left:.3rem}

.pm-section{margin-top:.8rem;padding:.6rem;background:var(--bg);border:1px solid var(--bg3)}
.pm-section h3{font-size:.75rem;color:var(--rl);margin-bottom:.3rem}
.pm-section textarea{min-height:50px;margin-bottom:.3rem}
.empty{text-align:center;padding:2rem;color:var(--cm);font-style:italic;font-family:var(--serif)}
</style>
<link href="https://fonts.googleapis.com/css2?family=Libre+Baskerville:ital@0;1&family=JetBrains+Mono:wght@400;600&display=swap" rel="stylesheet">
</head><body>
<div class="hdr">
<h1><span>Inquest</span></h1>
<div class="hdr-right">
<span>Active: <b id="sActive">-</b></span>
<span>MTTR: <b id="sMTTR">-</b></span>
<button class="btn btn-p" onclick="showDeclare()">Declare Incident</button>
</div>
</div>
<div class="main">
<div class="overview" id="overview"></div>
<div class="toolbar">
<select id="fStatus" onchange="loadIncs()"><option value="all">All</option><option value="active">Active</option><option value="investigating">Investigating</option><option value="identified">Identified</option><option value="monitoring">Monitoring</option><option value="resolved">Resolved</option></select>
<select id="fSev" onchange="loadIncs()"><option value="">Severity</option><option value="sev1">SEV1</option><option value="sev2">SEV2</option><option value="sev3">SEV3</option><option value="sev4">SEV4</option></select>
<input type="text" id="fSearch" placeholder="Search incidents..." onkeydown="if(event.key==='Enter')loadIncs()">
<button class="btn btn-p" onclick="loadIncs()">Search</button>
</div>
<div id="incList"></div>
</div>
<div id="modal"></div>

<script>
let incidents=[];
async function api(url,opts){const r=await fetch(url,opts);return r.json()}
function esc(s){return String(s||'').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;')}
function timeAgo(d){if(!d)return'';const s=Math.floor((Date.now()-new Date(d))/1e3);if(s<60)return s+'s ago';if(s<3600)return Math.floor(s/60)+'m ago';if(s<86400)return Math.floor(s/3600)+'h ago';return Math.floor(s/86400)+'d ago'}

async function init(){
  const sd=await api('/api/stats');
  document.getElementById('sActive').textContent=sd.active;
  document.getElementById('sMTTR').textContent=sd.mttr;
  const bySev=sd.by_severity||{};
  document.getElementById('overview').innerHTML=
    '<div class="stat"><b>'+sd.total+'</b>Total</div>'+
    '<div class="stat"><b>'+sd.active+'</b>Active</div>'+
    '<div class="stat"><b>'+sd.resolved+'</b>Resolved</div>'+
    '<div class="stat"><b>'+sd.postmortems+'</b>Postmortems</div>'+
    Object.entries(bySev).map(([k,v])=>'<div class="stat"><b>'+v+'</b>'+k.toUpperCase()+'</div>').join('');
  loadIncs();
}

async function loadIncs(){
  const p=new URLSearchParams();
  p.set('status',document.getElementById('fStatus').value);
  const sev=document.getElementById('fSev').value;if(sev)p.set('severity',sev);
  const q=document.getElementById('fSearch').value;if(q)p.set('search',q);
  const d=await api('/api/incidents?'+p);
  incidents=d.incidents||[];
  renderIncs();
}

function renderIncs(){
  const el=document.getElementById('incList');
  if(!incidents.length){el.innerHTML='<div class="empty">No incidents found.</div>';return}
  el.innerHTML=incidents.map(i=>{
    const services=(i.services||[]).map(s=>'<span style="color:var(--ll)">'+esc(s)+'</span>').join(', ');
    return '<div class="inc-card" onclick="showDetail(\''+i.id+'\')">'+
      '<div class="inc-top">'+
        '<span class="sev-badge sev-'+i.severity+'">'+i.severity.toUpperCase()+'</span>'+
        '<span class="status-badge st-'+i.status+'">'+i.status+'</span>'+
        '<span class="inc-title">'+esc(i.title)+'</span>'+
      '</div>'+
      '<div class="inc-meta">'+
        (i.lead?'<span>Lead: '+esc(i.lead)+'</span>':'')+
        '<span>Duration: '+i.duration+'</span>'+
        '<span>'+i.update_count+' updates</span>'+
        (i.has_postmortem?'<span style="color:var(--green)">Postmortem</span>':'')+
        (services?'<span>Services: '+services+'</span>':'')+
        '<span>'+timeAgo(i.started_at)+'</span>'+
      '</div></div>'
  }).join('')
}

async function showDetail(id){
  const [inc,tl,pm]=await Promise.all([api('/api/incidents/'+id),api('/api/incidents/'+id+'/timeline'),api('/api/incidents/'+id+'/postmortem')]);
  const timeline=(tl.timeline||[]).map(t=>'<div class="tl-item"><div class="tl-dot"></div><div class="tl-content">'+
    (t.author?'<b style="color:var(--rl)">'+esc(t.author)+'</b> ':'')+esc(t.message)+
    (t.status?'<span class="tl-status st-'+t.status+'">'+t.status+'</span>':'')+
    '<span class="tl-time">'+timeAgo(t.created_at)+'</span></div></div>').join('');
  const statusBtns=['investigating','identified','monitoring','resolved'].filter(s=>s!==inc.status).map(s=>
    '<button class="btn btn-'+(s==='resolved'?'s':'b')+'" onclick="setStatus(\''+id+'\',\''+s+'\')">'+s+'</button>').join(' ');
  const postmortem=pm&&pm.postmortem?pm:null;
  document.getElementById('modal').innerHTML='<div class="modal-bg" onclick="if(event.target===this)closeModal()"><div class="modal">'+
    '<div style="display:flex;justify-content:space-between;align-items:flex-start">'+
      '<h2><span class="sev-badge sev-'+inc.severity+'">'+inc.severity.toUpperCase()+'</span> '+esc(inc.title)+'</h2>'+
      '<button class="btn btn-d" onclick="if(confirm(\'Delete?\'))delInc(\''+id+'\')">Del</button>'+
    '</div>'+
    '<div style="display:flex;gap:.5rem;flex-wrap:wrap;margin:.5rem 0">'+
      '<span class="status-badge st-'+inc.status+'" style="font-size:.7rem">'+inc.status+'</span>'+
      statusBtns+
    '</div>'+
    '<div style="font-size:.7rem;color:var(--leather);margin:.5rem 0;display:flex;gap:1rem;flex-wrap:wrap">'+
      (inc.lead?'<span>Lead: '+esc(inc.lead)+'</span>':'')+
      '<span>Duration: '+inc.duration+'</span>'+
      '<span>Started: '+timeAgo(inc.started_at)+'</span>'+
      (inc.services.length?'<span>Services: '+inc.services.map(s=>esc(s)).join(', ')+'</span>':'')+
    '</div>'+
    (inc.summary?'<div style="padding:.5rem;background:var(--bg);border:1px solid var(--bg3);font-size:.78rem;color:var(--cd);margin:.5rem 0">'+esc(inc.summary)+'</div>':'')+
    '<div style="font-size:.7rem;color:var(--leather);margin:.8rem 0 .3rem">Timeline ('+tl.timeline.length+')</div>'+
    timeline+
    '<div style="margin-top:.6rem;display:flex;gap:.3rem;align-items:flex-end">'+
      '<div style="flex:1"><input type="text" id="updAuthor" placeholder="Name" style="font-size:.72rem;margin-bottom:.3rem"><input type="text" id="updMsg" placeholder="Status update..." style="font-size:.72rem"></div>'+
      '<button class="btn btn-p" onclick="postUpdate(\''+id+'\')">Update</button>'+
    '</div>'+
    '<div style="font-size:.7rem;color:var(--leather);margin:1rem 0 .3rem">Postmortem</div>'+
    (postmortem?renderPM(id,postmortem):'<button class="btn btn-p" onclick="createPM(\''+id+'\')">Start Postmortem</button>')+
  '</div></div>'
}

function renderPM(id,pm){
  return '<div class="pm-section"><h3>What Happened</h3><div style="font-size:.78rem;color:var(--cd);white-space:pre-wrap">'+esc(pm.what_happened||'(empty)')+'</div></div>'+
    '<div class="pm-section"><h3>Root Cause</h3><div style="font-size:.78rem;color:var(--cd);white-space:pre-wrap">'+esc(pm.root_cause||'(empty)')+'</div></div>'+
    '<div class="pm-section"><h3>Action Items</h3><div style="font-size:.78rem;color:var(--cd);white-space:pre-wrap">'+esc(pm.action_items||'(empty)')+'</div></div>'+
    '<div class="pm-section"><h3>Lessons Learned</h3><div style="font-size:.78rem;color:var(--cd);white-space:pre-wrap">'+esc(pm.lessons||'(empty)')+'</div></div>'+
    '<button class="btn btn-p" style="margin-top:.5rem" onclick="editPM(\''+id+'\')">Edit Postmortem</button>'
}

async function setStatus(id,status){
  await api('/api/incidents/'+id+'/status',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({status,author:'dashboard'})});
  showDetail(id);loadIncs();init()
}

async function postUpdate(id){
  const author=document.getElementById('updAuthor').value||'dashboard';
  const message=document.getElementById('updMsg').value;
  if(!message)return;
  await api('/api/incidents/'+id+'/timeline',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({author,message})});
  showDetail(id)
}

async function delInc(id){await api('/api/incidents/'+id,{method:'DELETE'});closeModal();loadIncs();init()}

async function createPM(id){
  await api('/api/incidents/'+id+'/postmortem',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({what_happened:'',root_cause:'',action_items:'',lessons:''})});
  editPM(id)
}

async function editPM(id){
  const pm=await api('/api/incidents/'+id+'/postmortem');
  document.getElementById('modal').innerHTML='<div class="modal-bg" onclick="if(event.target===this)closeModal()"><div class="modal">'+
    '<h2>Edit Postmortem</h2>'+
    '<label class="fl">What Happened</label><textarea id="pm-what" rows="3">'+esc(pm.what_happened||'')+'</textarea>'+
    '<label class="fl">Root Cause</label><textarea id="pm-root" rows="3">'+esc(pm.root_cause||'')+'</textarea>'+
    '<label class="fl">Action Items</label><textarea id="pm-action" rows="3">'+esc(pm.action_items||'')+'</textarea>'+
    '<label class="fl">Lessons Learned</label><textarea id="pm-lessons" rows="3">'+esc(pm.lessons||'')+'</textarea>'+
    '<div style="display:flex;gap:.5rem;margin-top:1rem"><button class="btn btn-p" onclick="savePM(\''+id+'\')">Save</button><button class="btn btn-d" onclick="showDetail(\''+id+'\')">Cancel</button></div>'+
  '</div></div>'
}

async function savePM(id){
  const body={what_happened:document.getElementById('pm-what').value,root_cause:document.getElementById('pm-root').value,action_items:document.getElementById('pm-action').value,lessons:document.getElementById('pm-lessons').value};
  await api('/api/incidents/'+id+'/postmortem',{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
  showDetail(id)
}

function showDeclare(){
  document.getElementById('modal').innerHTML='<div class="modal-bg" onclick="if(event.target===this)closeModal()"><div class="modal">'+
    '<h2>Declare Incident</h2>'+
    '<label class="fl">Title</label><input type="text" id="di-title" placeholder="API is returning 500s">'+
    '<div class="form-row">'+
      '<div><label class="fl">Severity</label><select id="di-sev"><option value="sev1">SEV1 - Critical</option><option value="sev2">SEV2 - High</option><option value="sev3" selected>SEV3 - Medium</option><option value="sev4">SEV4 - Low</option></select></div>'+
      '<div><label class="fl">Lead</label><input type="text" id="di-lead" placeholder="Incident commander"></div>'+
    '</div>'+
    '<label class="fl">Summary</label><textarea id="di-summary" rows="2" placeholder="Brief description of the issue"></textarea>'+
    '<label class="fl">Affected Services (comma-separated)</label><input type="text" id="di-services" placeholder="api, database, proxy">'+
    '<div style="display:flex;gap:.5rem;margin-top:1rem"><button class="btn btn-p" onclick="saveDeclare()">Declare</button><button class="btn btn-d" onclick="closeModal()">Cancel</button></div>'+
  '</div></div>'
}

async function saveDeclare(){
  const services=(document.getElementById('di-services').value||'').split(',').map(s=>s.trim()).filter(Boolean);
  const body={title:document.getElementById('di-title').value,severity:document.getElementById('di-sev').value,lead:document.getElementById('di-lead').value,summary:document.getElementById('di-summary').value,services};
  if(!body.title){alert('Title required');return}
  await api('/api/incidents',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
  closeModal();loadIncs();init()
}

function closeModal(){document.getElementById('modal').innerHTML=''}
init();
</script></body></html>`
