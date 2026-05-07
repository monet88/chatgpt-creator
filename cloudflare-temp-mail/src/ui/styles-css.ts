export const uiStyles = `:root {
  --ink: #111111;
  --paper: #fff8df;
  --yellow: #ffe14d;
  --cyan: #41d9ff;
  --pink: #ff5da2;
  --green: #6cff8d;
  --shadow: 8px 8px 0 var(--ink);
  font-family: Inter, ui-sans-serif, system-ui, sans-serif;
}

* { box-sizing: border-box; }
body {
  margin: 0;
  color: var(--ink);
  background:
    linear-gradient(90deg, rgba(17,17,17,.08) 1px, transparent 1px),
    linear-gradient(rgba(17,17,17,.08) 1px, transparent 1px),
    var(--paper);
  background-size: 28px 28px;
}
button, input, select { font: inherit; }
button, select, input {
  border: 3px solid var(--ink);
  color: var(--ink);
  background: white;
  min-height: 44px;
}
button {
  cursor: pointer;
  font-weight: 900;
  text-transform: uppercase;
  box-shadow: 4px 4px 0 var(--ink);
}
button:hover { transform: translate(-1px, -1px); box-shadow: 5px 5px 0 var(--ink); }
button:active { transform: translate(2px, 2px); box-shadow: 2px 2px 0 var(--ink); }
button:focus-visible, input:focus-visible, select:focus-visible { outline: 4px solid var(--pink); outline-offset: 3px; }
.shell {
  width: min(1120px, calc(100% - 32px));
  margin: 32px auto;
  display: grid;
  grid-template-columns: 1.5fr .8fr;
  gap: 24px;
}
.hero-card, .status-card, .inbox-card, .modal-card {
  border: 4px solid var(--ink);
  background: white;
  box-shadow: var(--shadow);
  padding: clamp(20px, 4vw, 40px);
}
.hero-card { background: linear-gradient(135deg, var(--yellow), white 55%); }
.hero-card h1 { font-size: clamp(2.5rem, 8vw, 6rem); line-height: .9; margin: 0 0 28px; letter-spacing: -0.08em; max-width: 760px; }
.eyebrow { font-family: ui-monospace, SFMono-Regular, monospace; font-weight: 900; letter-spacing: .18em; margin: 0 0 12px; }
.mailbox-panel { display: grid; gap: 12px; max-width: 720px; }
.email-row { display: grid; grid-template-columns: 1fr auto; gap: 12px; }
.email-row input { width: 100%; padding: 0 14px; font-family: ui-monospace, SFMono-Regular, monospace; font-weight: 800; }
.action-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 12px; margin-top: 8px; }
.danger { background: var(--pink); }
.status-card { background: var(--cyan); display: grid; align-content: start; gap: 24px; }
.badge { display: inline-block; width: fit-content; border: 3px solid var(--ink); background: var(--green); padding: 8px 12px; font-weight: 1000; box-shadow: 4px 4px 0 var(--ink); }
.otp-box { width: 100%; padding: 20px; background: white; font-size: clamp(1.5rem, 5vw, 3rem); font-family: ui-monospace, SFMono-Regular, monospace; }
.inbox-card { grid-column: 1 / -1; background: white; }
.section-head { display: flex; justify-content: space-between; align-items: end; gap: 16px; margin-bottom: 16px; }
h2 { margin: 0; font-size: clamp(2rem, 5vw, 4rem); letter-spacing: -0.06em; }
.table-wrap { overflow-x: auto; }
table { width: 100%; border-collapse: collapse; min-width: 680px; }
th, td { border: 3px solid var(--ink); padding: 12px; text-align: left; vertical-align: top; }
td { line-height: 1.45; }
th { background: var(--yellow); text-transform: uppercase; }
.inline-error { min-height: 1.2em; color: #9b003f; font-weight: 900; }
dialog { width: min(760px, calc(100% - 28px)); border: 0; background: transparent; padding: 0; }
dialog::backdrop { background: rgba(0,0,0,.55); }
.modal-card { position: relative; }
.close-button { float: right; background: var(--yellow); }
pre { white-space: pre-wrap; font-family: ui-monospace, SFMono-Regular, monospace; background: var(--paper); border: 3px solid var(--ink); padding: 16px; max-height: 55vh; overflow: auto; }
#toast { position: fixed; right: 16px; bottom: 16px; background: var(--ink); color: white; padding: 12px 16px; font-weight: 900; opacity: 0; transform: translateY(8px); transition: 160ms ease; }
#toast.show { opacity: 1; transform: translateY(0); }
@media (max-width: 760px) {
  .shell { grid-template-columns: 1fr; width: min(100% - 20px, 1120px); margin: 16px auto; }
  .email-row, .action-grid { grid-template-columns: 1fr; }
  .hero-card, .status-card, .inbox-card { box-shadow: 5px 5px 0 var(--ink); }
}
`;
