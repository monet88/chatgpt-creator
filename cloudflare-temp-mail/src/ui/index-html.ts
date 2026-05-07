export const uiHtml = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Cloudflare Temp Mail</title>
    <link rel="stylesheet" href="/assets/styles.css" />
  </head>
  <body>
    <main class="shell">
      <section class="hero-card" aria-labelledby="app-title">
        <p class="eyebrow">WORKER MAILBOX</p>
        <h1 id="app-title">Temporary inbox. Brutal speed.</h1>
        <div class="mailbox-panel">
          <label for="domain-select">Domain</label>
          <select id="domain-select"></select>
          <label for="email-output">Address</label>
          <div class="email-row">
            <input id="email-output" readonly value="Generate an address" />
            <button id="copy-email" type="button">Copy</button>
          </div>
          <div class="action-grid">
            <button id="generate-email" type="button">Generate</button>
            <button id="refresh-inbox" type="button">Refresh</button>
            <button id="delete-mailbox" type="button" class="danger">Delete all</button>
          </div>
        </div>
      </section>

      <section class="status-card" aria-live="polite">
        <span id="status-badge" class="badge">READY</span>
        <div>
          <p class="eyebrow">LATEST OTP</p>
          <button id="copy-otp" type="button" class="otp-box">No OTP yet</button>
        </div>
      </section>

      <section class="inbox-card" aria-labelledby="inbox-title">
        <div class="section-head">
          <h2 id="inbox-title">Inbox</h2>
          <span id="message-count">0 messages</span>
        </div>
        <p id="inline-error" class="inline-error" role="alert"></p>
        <div class="table-wrap">
          <table>
            <thead>
              <tr><th>From</th><th>Subject</th><th>Date</th><th>Action</th></tr>
            </thead>
            <tbody id="inbox-body">
              <tr><td colspan="4">Generate an address to start.</td></tr>
            </tbody>
          </table>
        </div>
      </section>
    </main>

    <dialog id="message-modal" aria-labelledby="modal-subject" aria-describedby="modal-body">
      <article class="modal-card">
        <button id="close-modal" type="button" class="close-button">Close</button>
        <p id="modal-from" class="eyebrow"></p>
        <h2 id="modal-subject"></h2>
        <pre id="modal-body"></pre>
      </article>
    </dialog>

    <div id="toast" role="status" aria-live="polite"></div>
    <script type="module" src="/assets/app.js"></script>
  </body>
</html>`;
