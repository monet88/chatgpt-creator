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
.message-pager{align-items:center;justify-content:space-between;gap:16px;margin:16px 24px;
  background:#f8fafc;color:#0f172a;border-radius:8px;padding:10px 14px;flex-wrap:wrap}
.page-size-label{display:flex;align-items:center;gap:8px;font-weight:600;color:#020617}
.page-size-label select{background:#fff;color:#020617;border:1px solid #d8dee6;border-radius:4px;padding:6px 28px 6px 8px}
.pager-status{font-weight:600;color:#020617}
.pager-buttons{display:flex;align-items:center;gap:6px}
.pager-btn{border:1px solid #d8dee6;background:#fff;color:#94a3b8;border-radius:4px;padding:7px 11px;cursor:pointer}
.pager-btn:not(:disabled):hover{color:#0f172a;border-color:#0ea5e9}
.pager-btn:disabled{opacity:.55;cursor:not-allowed}
.pager-page{background:#0ea5e9;color:#fff;border-radius:5px;padding:8px 12px;font-weight:700}
.message-table-wrap{margin:0 24px;background:#f8fafc;overflow:hidden}
.message-table{width:100%;border-collapse:collapse;color:#0f172a}
.message-table th{padding:13px 15px;text-align:left;font-size:14px;font-weight:700;color:#fff;background:#304b60}
.message-table td{padding:13px 15px;border-bottom:1px solid #e5eaf0;font-size:14px;vertical-align:middle;background:#f8fafc}
.message-table tr:hover td,.message-table tr.selected td{background:#eef6ff}
.message-table .btn-read{background:#1899d6;color:#fff;border:none;border-radius:5px;padding:7px 12px;font-size:14px;cursor:pointer}
.message-table .btn-read:hover{background:#0b89c5}
.message-table .btn-del{background:#ef4444;border:none;color:#fff;border-radius:5px;padding:7px 12px;font-size:14px;cursor:pointer;margin-left:6px}
.message-table .btn-del:hover{background:#dc2626}
.from-cell{max-width:220px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.subject-cell{max-width:320px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.date-cell{white-space:nowrap;color:#0f172a}
.actions-cell{white-space:nowrap}

/* Inline detail */
.message-detail{margin:22px 38px 40px;background:#fbfdff;color:#0f172a;border:1px solid #d7e1ec;
  border-radius:8px;padding:38px 20px 20px;box-shadow:0 10px 24px rgba(15,23,42,.08)}
.message-detail h3{font-size:24px;line-height:1.2;color:#28445f;margin-bottom:6px}
.message-detail p{font-size:16px;line-height:1.45;margin:8px 0;color:#0f172a}
.message-detail strong{font-weight:700}
.detail-divider{height:2px;background:#0ea5e9;margin:4px 0 10px}
.detail-body{margin-top:10px;border:1px solid #d7e1ec;border-radius:5px;padding:16px;background:#fff;overflow:hidden}
.detail-text{font-family:Inter,ui-sans-serif,system-ui,sans-serif;font-size:16px;line-height:1.6;
  color:#0f172a;border:1px solid #d7e1ec;border-radius:5px;padding:16px;white-space:pre-wrap;overflow:auto;max-height:58vh;display:block}
.detail-iframe{width:100%;height:55vh;border:1px solid #d7e1ec;background:#fff;display:block;border-radius:5px}

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
