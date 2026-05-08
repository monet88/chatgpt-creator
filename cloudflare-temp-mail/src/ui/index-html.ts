export const uiHtml = `<!doctype html>
<html lang="vi">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>TempMail — Email tạm thời miễn phí</title>
  <link rel="stylesheet" href="/assets/styles.css" />
</head>
<body>
  <nav class="navbar">
    <a class="brand" href="/">
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="4" width="20" height="16" rx="2"/><path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7"/></svg>
      <span>TempMail</span>
    </a>
    <div class="nav-links">
      <a href="/domain" class="nav-link">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20"/><path d="M2 12h20"/></svg>
        Thêm Domain
      </a>
      <a href="/api" class="nav-link">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>
        API Docs
      </a>
    </div>
    <span class="realtime-dot">● Realtime</span>
  </nav>

  <div class="app-layout">
    <aside class="sidebar">
      <div class="sidebar-section">
        <p class="sidebar-label">EMAIL HIỆN TẠI</p>
        <p id="current-email" class="current-email">Chưa có địa chỉ</p>
        <a id="url-email" class="url-email" href="#" title="Link chia sẻ hộp thư này"></a>
        <button id="copy-btn" class="btn btn-primary">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
          Sao chép địa chỉ
        </button>
      </div>

      <div id="otp-section" class="sidebar-section otp-section" style="display:none">
        <p class="sidebar-label">MÃ 2FA / OTP MỚI NHẤT</p>
        <div class="otp-row">
          <code id="otp-code" class="otp-code">—</code>
          <button id="copy-otp-btn" class="btn btn-sm btn-primary">Sao chép mã</button>
        </div>
      </div>

      <div class="sidebar-section">
        <p class="sidebar-label">Tuỳ chỉnh địa chỉ</p>
        <div class="custom-row">
          <input id="username-input" type="text" placeholder="username" autocomplete="off" spellcheck="false" />
          <span class="at-sign">@</span>
          <select id="domain-select"></select>
        </div>
        <button id="generate-btn" class="btn btn-accent">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 2v6h-6"/><path d="M3 12a9 9 0 0 1 15-6.7L21 8"/><path d="M3 22v-6h6"/><path d="M21 12a9 9 0 0 1-15 6.7L3 16"/></svg>
          Tạo Email Ngẫu Nhiên
        </button>
      </div>

      <div class="stats-grid">
        <div class="stat-card">
          <p id="inbox-count" class="stat-num">0</p>
          <p class="stat-label">Trong hộp thư</p>
        </div>
        <div class="stat-card">
          <p id="domain-count" class="stat-num">0</p>
          <p class="stat-label">Domains hoạt động</p>
        </div>
      </div>

      <div class="sidebar-section howto">
        <p class="sidebar-label">CÁCH SỬ DỤNG</p>
        <ol>
          <li>Sao chép địa chỉ email ở trên</li>
          <li>Dùng để đăng ký dịch vụ bất kỳ</li>
          <li>Email sẽ xuất hiện ngay tại đây</li>
        </ol>
      </div>
    </aside>

    <main class="inbox-area">
      <div class="inbox-header">
        <div class="inbox-title-row">
          <h2>Hộp thư đến</h2>
          <code id="email-badge" class="email-badge"></code>
          <span id="unread-count" class="unread-badge">0</span>
        </div>
        <div class="inbox-toolbar">
          <label class="toggle-label">
            <input id="auto-refresh-toggle" type="checkbox" checked />
            <span>Auto-refresh</span>
          </label>
          <button id="delete-all-btn" class="btn btn-danger btn-sm">Xóa tất cả</button>
        </div>
      </div>

      <div id="empty-state" class="empty-state">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" opacity=".3"><rect x="2" y="4" width="20" height="16" rx="2"/><path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7"/></svg>
        <h3>Hộp thư đang chờ email</h3>
        <p>Sao chép địa chỉ email ở bên trái và dùng để đăng ký.<br>Email sẽ xuất hiện ngay tại đây.</p>
      </div>

      <table id="message-table" class="message-table" style="display:none">
        <thead><tr><th>TỪ</th><th>TIÊU ĐỀ</th><th>NGÀY</th><th>THAO TÁC</th></tr></thead>
        <tbody id="inbox-body"></tbody>
      </table>
    </main>
  </div>

  <div id="inline-error" class="inline-error" role="alert"></div>

  <dialog id="msg-modal">
    <div class="modal-card">
      <button id="close-modal" class="modal-close">✕</button>
      <p id="modal-from" class="modal-meta"></p>
      <h3 id="modal-subject" class="modal-subject"></h3>
      <pre id="modal-body" class="modal-body"></pre>
    </div>
  </dialog>

  <div id="toast" class="toast" role="status" aria-live="polite"></div>
  <script src="/assets/app.js"></script>
</body>
</html>`;
