export const uiStyles = `
*{box-sizing:border-box;margin:0;padding:0}
:root{
  --bg:#09090b;--surface:#18181b;--surface2:#27272a;
  --border:rgba(255,255,255,0.08);--border2:rgba(255,255,255,0.15);
  --text:#fafafa;--muted:#a1a1aa;--accent:#8b5cf6;--accent-hover:#7c3aed;
  --green:#22c55e;--red:#ef4444;--red-hover:#dc2626;
  font-family:Inter,ui-sans-serif,system-ui,sans-serif;
  font-size:14px;color-scheme:dark;
}
body{background:var(--bg);color:var(--text);min-height:100vh;display:flex;flex-direction:column}
a{color:inherit;text-decoration:none}
button,input,select{font:inherit;color:inherit}

/* Navbar */
.navbar{height:56px;background:var(--surface);border-bottom:1px solid var(--border);
  display:flex;align-items:center;padding:0 24px;gap:24px;position:sticky;top:0;z-index:10}
.brand{display:flex;align-items:center;gap:8px;font-weight:700;font-size:15px;color:var(--accent)}
.brand svg{color:var(--accent)}
.nav-links{display:flex;gap:4px;margin-left:8px}
.nav-link{display:flex;align-items:center;gap:6px;padding:6px 12px;border-radius:6px;
  color:var(--muted);font-size:13px;transition:all 120ms}
.nav-link:hover{color:var(--text);background:var(--surface2)}
.realtime-dot{margin-left:auto;color:var(--green);font-size:12px;font-weight:500}

/* Layout */
.app-layout{display:grid;grid-template-columns:300px 1fr;flex:1;overflow:hidden;height:calc(100vh - 56px)}

/* Sidebar */
.sidebar{background:var(--surface);border-right:1px solid var(--border);
  overflow-y:auto;padding:20px 16px;display:flex;flex-direction:column;gap:20px}
.sidebar-section{display:flex;flex-direction:column;gap:10px}
.sidebar-label{font-size:11px;font-weight:600;letter-spacing:.1em;color:var(--muted)}
.current-email{font-family:ui-monospace,monospace;font-size:13px;font-weight:600;
  word-break:break-all;color:var(--text);background:var(--surface2);
  padding:10px 12px;border-radius:6px;border:1px solid var(--border);min-height:40px}
.custom-row{display:grid;grid-template-columns:1fr auto 1fr;align-items:center;gap:4px}
.custom-row input,.custom-row select{background:var(--surface2);border:1px solid var(--border);
  border-radius:6px;padding:8px 10px;font-size:13px;width:100%;outline:none;transition:border 120ms}
.custom-row input:focus,.custom-row select:focus{border-color:var(--accent)}
.custom-row select{padding-right:6px;cursor:pointer}
.at-sign{color:var(--muted);font-size:13px;text-align:center;flex-shrink:0}
.stats-grid{display:grid;grid-template-columns:1fr 1fr;gap:8px}
.stat-card{background:var(--surface2);border:1px solid var(--border);border-radius:8px;
  padding:14px 12px;text-align:center}
.stat-num{font-size:24px;font-weight:700;color:var(--text)}
.stat-label{font-size:11px;color:var(--muted);margin-top:2px}
.howto ol{padding-left:18px;display:flex;flex-direction:column;gap:6px;color:var(--muted);line-height:1.5}
.url-email{font-size:11px;color:var(--accent);word-break:break-all;line-height:1.4;
  text-decoration:none;display:block;min-height:0}
.url-email:empty{display:none}
.url-email:hover{text-decoration:underline}
.otp-section{background:rgba(139,92,246,.08);border:1px solid rgba(139,92,246,.2);
  border-radius:8px;padding:12px}
.otp-row{display:flex;align-items:center;gap:8px;margin-top:4px}
.otp-code{font-family:ui-monospace,monospace;font-size:22px;font-weight:700;
  color:var(--text);letter-spacing:.15em;flex:1}

/* Buttons */
.btn{display:inline-flex;align-items:center;justify-content:center;gap:6px;
  border:none;border-radius:6px;padding:9px 14px;font-size:13px;font-weight:500;
  cursor:pointer;transition:all 120ms;white-space:nowrap}
.btn-primary{background:var(--surface2);border:1px solid var(--border2);color:var(--text)}
.btn-primary:hover{background:var(--border2)}
.btn-accent{background:var(--accent);color:#fff}
.btn-accent:hover{background:var(--accent-hover)}
.btn-danger{background:transparent;border:1px solid var(--border2);color:var(--red)}
.btn-danger:hover{background:var(--red);color:#fff;border-color:var(--red)}
.btn-ghost{background:transparent;border:1px solid var(--border);color:var(--muted);padding:7px 10px}
.btn-ghost:hover{color:var(--text);border-color:var(--border2)}
.btn-sm{padding:6px 10px;font-size:12px}
.btn-full{width:100%}

/* Inbox */
.inbox-area{overflow-y:auto;display:flex;flex-direction:column;background:var(--bg)}
.inbox-header{padding:20px 24px 16px;border-bottom:1px solid var(--border);
  display:flex;align-items:center;justify-content:space-between;gap:12px;flex-wrap:wrap;
  position:sticky;top:0;background:var(--bg);z-index:5}
.inbox-title-row{display:flex;align-items:center;gap:10px}
.inbox-title-row h2{font-size:18px;font-weight:600}
.email-badge{font-size:12px;color:var(--muted);background:var(--surface2);
  padding:3px 8px;border-radius:4px;font-family:ui-monospace,monospace}
.unread-badge{background:var(--accent);color:#fff;font-size:11px;font-weight:600;
  padding:2px 7px;border-radius:10px;min-width:20px;text-align:center}
.inbox-toolbar{display:flex;align-items:center;gap:8px}
.toggle-label{display:flex;align-items:center;gap:6px;cursor:pointer;
  color:var(--muted);font-size:12px;user-select:none}
.toggle-label input{accent-color:var(--green);width:14px;height:14px}
.toggle-label:has(input:checked) span{color:var(--green)}

/* Empty state */
.empty-state{flex:1;display:flex;flex-direction:column;align-items:center;justify-content:center;
  gap:12px;padding:40px;text-align:center}
.empty-state h3{font-size:16px;font-weight:600;color:var(--muted)}
.empty-state p{color:var(--muted);font-size:13px;line-height:1.6;max-width:360px}

/* Table */
.message-table{width:100%;border-collapse:collapse}
.message-table th{padding:11px 16px;text-align:left;font-size:11px;font-weight:600;
  letter-spacing:.08em;color:var(--muted);border-bottom:1px solid var(--border);
  background:var(--surface);position:sticky;top:0}
.message-table td{padding:13px 16px;border-bottom:1px solid var(--border);
  font-size:13px;vertical-align:middle}
.message-table tr:hover td{background:var(--surface)}
.message-table .btn-read{background:var(--accent);color:#fff;border:none;
  border-radius:4px;padding:4px 10px;font-size:12px;cursor:pointer}
.message-table .btn-del{background:transparent;border:1px solid var(--border2);
  color:var(--muted);border-radius:4px;padding:4px 8px;font-size:12px;cursor:pointer;margin-left:4px}
.message-table .btn-del:hover{background:var(--red);color:#fff;border-color:var(--red)}
.from-cell{max-width:180px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.subject-cell{max-width:280px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.date-cell{white-space:nowrap;color:var(--muted);font-size:12px}

/* Modal */
dialog{border:none;background:transparent;padding:0;max-width:720px;width:calc(100% - 32px)}
dialog::backdrop{background:rgba(0,0,0,.7);backdrop-filter:blur(4px)}
.modal-card{background:var(--surface);border:1px solid var(--border2);border-radius:12px;
  padding:28px;position:relative;max-height:85vh;overflow:auto}
.modal-close{position:absolute;top:16px;right:16px;background:var(--surface2);
  border:1px solid var(--border);border-radius:6px;color:var(--muted);
  padding:5px 9px;cursor:pointer;font-size:13px}
.modal-close:hover{color:var(--text)}
.modal-meta{font-size:12px;color:var(--muted);margin-bottom:6px}
.modal-subject{font-size:18px;font-weight:600;margin-bottom:16px;padding-right:40px}
.modal-body{background:var(--surface2);border:1px solid var(--border);border-radius:6px;
  overflow:hidden;max-height:58vh}
.modal-text{font-family:ui-monospace,monospace;font-size:12px;line-height:1.6;
  padding:16px;white-space:pre-wrap;overflow:auto;max-height:58vh;color:var(--muted);display:block}
.modal-iframe{width:100%;height:55vh;border:none;background:#fff;display:block;border-radius:6px}

/* Toast & Error */
.toast{position:fixed;bottom:20px;right:20px;background:var(--surface);
  border:1px solid var(--border2);color:var(--text);padding:10px 16px;border-radius:8px;
  font-size:13px;opacity:0;transform:translateY(6px);transition:all 160ms;pointer-events:none;z-index:100}
.toast.show{opacity:1;transform:translateY(0)}
.inline-error{position:fixed;top:64px;left:50%;transform:translateX(-50%);
  background:var(--red);color:#fff;padding:8px 16px;border-radius:6px;
  font-size:13px;opacity:0;transition:opacity 160ms;pointer-events:none;z-index:20}
.inline-error.show{opacity:1}

@media(max-width:768px){
  .app-layout{grid-template-columns:1fr;height:auto}
  .sidebar{border-right:none;border-bottom:1px solid var(--border)}
  .inbox-area{min-height:60vh}
  .nav-links .nav-link span{display:none}
}
`;
