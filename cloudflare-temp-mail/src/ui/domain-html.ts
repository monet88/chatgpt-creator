export const domainHtml = `<!doctype html>
<html lang="vi">
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width,initial-scale=1"/>
  <title>Thêm Domain — TempMail</title>
  <link rel="stylesheet" href="/assets/styles.css"/>
  <style>
    .page{max-width:760px;margin:0 auto;padding:48px 24px}
    .page-tag{display:inline-flex;align-items:center;gap:6px;font-size:12px;font-weight:600;
      color:var(--accent);background:rgba(139,92,246,.12);border:1px solid rgba(139,92,246,.3);
      border-radius:20px;padding:4px 12px;margin-bottom:20px}
    .page h1{font-size:42px;font-weight:700;line-height:1.1;margin-bottom:12px}
    .page>p{color:var(--muted);font-size:15px;line-height:1.6;margin-bottom:40px}
    .step{display:flex;gap:20px;margin-bottom:32px}
    .step-num{width:36px;height:36px;border-radius:50%;background:var(--accent);
      color:#fff;font-weight:700;font-size:14px;display:flex;align-items:center;
      justify-content:center;flex-shrink:0;margin-top:2px}
    .step-content h3{font-size:16px;font-weight:600;margin-bottom:8px}
    .step-content p{color:var(--muted);font-size:14px;line-height:1.6;margin-bottom:12px}
    .callout{background:rgba(234,179,8,.08);border:1px solid rgba(234,179,8,.25);
      border-radius:8px;padding:12px 16px;font-size:13px;color:#fbbf24;margin-top:8px}
    .code-block{background:var(--surface2);border:1px solid var(--border);border-radius:8px;
      padding:16px;font-family:ui-monospace,monospace;font-size:13px;margin-top:8px;overflow-x:auto}
    .code-block table{width:100%;border-collapse:collapse}
    .code-block td{padding:4px 12px 4px 0;vertical-align:top;white-space:nowrap;color:var(--muted)}
    .code-block td:first-child{color:var(--accent);font-weight:600}
    .code-block td:last-child{color:var(--text)}
    .back-link{display:inline-flex;align-items:center;gap:6px;color:var(--muted);
      font-size:13px;margin-bottom:32px}
    .back-link:hover{color:var(--text)}
  </style>
</head>
<body>
  <nav class="navbar">
    <a class="brand" href="/">
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="4" width="20" height="16" rx="2"/><path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7"/></svg>
      TempMail
    </a>
    <div class="nav-links">
      <a href="/" class="nav-link">Trang chủ</a>
      <a href="/domain" class="nav-link" style="color:var(--text)">Domain</a>
      <a href="/api" class="nav-link">API</a>
    </div>
  </nav>

  <div class="page">
    <a href="/" class="back-link">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="m15 18-6-6 6-6"/></svg>
      Về trang chủ
    </a>
    <div class="page-tag">
      <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 2a14.5 14.5 0 0 0 0 20"/><path d="M2 12h20"/></svg>
      Quản lý Domain
    </div>
    <h1>Thêm Domain<br>của bạn</h1>
    <p>Kết nối domain riêng vào hệ thống TempMail để tạo địa chỉ email tạm thời với tên miền của bạn. Chỉ cần cấu hình Email Routing trên Cloudflare, toàn bộ mail sẽ được xử lý tự động.</p>

    <div class="step">
      <div class="step-num">1</div>
      <div class="step-content">
        <h3>Chuyển DNS domain về Cloudflare</h3>
        <p>Domain của bạn phải dùng Cloudflare làm DNS. Nếu chưa, hãy vào nhà đăng ký domain và đổi nameserver về Cloudflare. Sau đó thêm domain vào tài khoản Cloudflare tại <strong>dash.cloudflare.com</strong>.</p>
        <div class="callout">Nếu domain đang có email thật, hãy dùng subdomain (vd: <strong>tmp.yourdomain.com</strong>) để tránh ảnh hưởng đến email hiện tại.</div>
      </div>
    </div>

    <div class="step">
      <div class="step-num">2</div>
      <div class="step-content">
        <h3>Bật Email Routing và tạo Catch-all rule</h3>
        <p>Vào <strong>Cloudflare Dashboard → yourdomain.com → Email → Email Routing</strong>, bật Email Routing. Sau đó vào tab <strong>Routing Rules → Catch-all address</strong>, chọn action <strong>Send to a Worker</strong> và chọn worker <strong>cloudflare-temp-mail</strong>.</p>
        <div class="callout">Catch-all rule sẽ forward <em>mọi email</em> gửi đến <strong>*@yourdomain.com</strong> vào worker để xử lý.</div>
      </div>
    </div>

    <div class="step">
      <div class="step-num">3</div>
      <div class="step-content">
        <h3>Thêm domain vào database</h3>
        <p>Chạy lệnh sau để kích hoạt domain trong hệ thống:</p>
        <div class="code-block">
          npx wrangler d1 execute cloudflare-temp-mail --remote \\<br>
          &nbsp;&nbsp;--command "INSERT OR IGNORE INTO domains(domain, enabled) VALUES ('yourdomain.com', 1);"
        </div>
      </div>
    </div>

    <div class="step">
      <div class="step-num">4</div>
      <div class="step-content">
        <h3>Cập nhật ENABLED_DOMAINS</h3>
        <p>Thêm domain mới vào biến <code>ENABLED_DOMAINS</code> trong <strong>wrangler.toml</strong> (phân cách bằng dấu phẩy), sau đó deploy lại worker:</p>
        <div class="code-block"><table>
          <tr><td>ENABLED_DOMAINS</td><td>=</td><td>"yourdomain.com,domain2.com"</td></tr>
        </table></div>
        <div class="code-block">npx wrangler deploy</div>
      </div>
    </div>
  </div>
</body>
</html>`;
