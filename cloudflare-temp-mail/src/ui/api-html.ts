export const apiHtml = `<!doctype html>
<html lang="vi">
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width,initial-scale=1"/>
  <title>API Docs — TempMail</title>
  <link rel="stylesheet" href="/assets/styles.css"/>
  <style>
    .doc-wrap{max-width:900px;margin:0 auto;padding:0 24px 80px}
    .doc-tabs{position:sticky;top:56px;background:var(--bg);border-bottom:1px solid var(--border);
      display:flex;gap:0;z-index:8;overflow-x:auto}
    .doc-tab{padding:14px 20px;font-size:13px;font-weight:500;color:var(--muted);cursor:pointer;
      white-space:nowrap;border-bottom:2px solid transparent;background:none;border-top:none;
      border-left:none;border-right:none;transition:all 120ms}
    .doc-tab:hover{color:var(--text)}
    .doc-tab.active{color:var(--accent);border-bottom-color:var(--accent)}
    .doc-section{display:none;padding-top:36px}
    .doc-section.active{display:block}
    .section-intro{color:var(--muted);font-size:14px;line-height:1.7;margin-bottom:32px}
    .base-url-box{background:var(--surface);border:1px solid var(--border);border-radius:8px;
      padding:14px 18px;display:flex;align-items:center;gap:16px;margin:20px 0 32px;flex-wrap:wrap}
    .base-url-box .label{font-size:11px;font-weight:600;letter-spacing:.1em;color:var(--muted);white-space:nowrap}
    .base-url-box code{font-family:ui-monospace,monospace;font-size:13px;color:var(--accent);flex:1}
    .toc{background:var(--surface);border:1px solid var(--border);border-radius:8px;
      padding:20px 24px;margin-bottom:36px}
    .toc h3{font-size:14px;font-weight:600;margin-bottom:14px;color:var(--muted)}
    .toc ol{padding-left:20px;display:flex;flex-direction:column;gap:4px}
    .toc li{font-size:13px;color:var(--muted)}
    .toc a{color:var(--accent)}
    .toc a:hover{text-decoration:underline}
    .toc .sub{list-style:none;padding-left:12px;margin-top:4px;display:flex;flex-direction:column;gap:3px}
    .toc .sub a{color:var(--muted)}
    .ep-block{border:1px solid var(--border);border-radius:10px;margin-bottom:24px;overflow:hidden}
    .ep-title{border-left:3px solid var(--accent);padding:16px 20px;background:var(--surface)}
    .ep-title h3{font-size:16px;font-weight:600;margin-bottom:2px}
    .ep-line{display:flex;align-items:center;gap:10px;margin:10px 0 6px}
    .method{font-family:ui-monospace,monospace;font-size:11px;font-weight:700;
      padding:3px 9px;border-radius:4px;min-width:48px;text-align:center}
    .GET{background:rgba(34,197,94,.15);color:#4ade80}
    .POST{background:rgba(59,130,246,.15);color:#60a5fa}
    .DELETE{background:rgba(239,68,68,.15);color:#f87171}
    .ep-path{font-family:ui-monospace,monospace;font-size:13px}
    .ep-desc{font-size:13px;color:var(--muted)}
    .ep-body{padding:0 20px 20px}
    .param-table{width:100%;border-collapse:collapse;margin-top:14px}
    .param-table th{text-align:left;padding:8px 12px;font-size:11px;font-weight:600;
      letter-spacing:.08em;color:var(--muted);background:var(--surface2);border-bottom:1px solid var(--border)}
    .param-table td{padding:10px 12px;font-size:13px;border-bottom:1px solid var(--border);vertical-align:top}
    .param-table tr:last-child td{border-bottom:none}
    .param-name{font-family:ui-monospace,monospace;font-size:12px;background:var(--surface2);
      padding:2px 6px;border-radius:3px;color:var(--text)}
    .param-type{font-size:11px;color:var(--muted);font-style:italic;white-space:nowrap}
    .param-desc{color:var(--muted);font-size:13px}
    .required{color:#f87171;font-size:10px;font-weight:600;margin-left:4px}
    .section-label{font-size:11px;font-weight:600;letter-spacing:.1em;color:var(--muted);
      margin:16px 0 8px}
    .code-block{background:var(--surface2);border:1px solid var(--border);border-radius:6px;
      padding:14px 16px;font-family:ui-monospace,monospace;font-size:12px;line-height:1.6;
      overflow-x:auto;color:var(--muted);white-space:pre}
    .error-table{width:100%;border-collapse:collapse;margin-top:12px}
    .error-table th{text-align:left;padding:9px 14px;font-size:11px;font-weight:600;
      letter-spacing:.08em;color:var(--muted);background:var(--surface2);border-bottom:1px solid var(--border)}
    .error-table td{padding:10px 14px;font-size:13px;border-bottom:1px solid var(--border);color:var(--muted)}
    .error-table tr:last-child td{border-bottom:none}
    .note-box{background:rgba(234,179,8,.06);border:1px solid rgba(234,179,8,.2);
      border-radius:8px;padding:14px 16px;font-size:13px;color:#fbbf24;line-height:1.6;margin-bottom:12px}
    .copy-btn{background:var(--accent);color:#fff;border:none;border-radius:5px;
      padding:4px 10px;font-size:11px;cursor:pointer;float:right;margin-top:-2px}
    .copy-btn:hover{background:var(--accent-hover)}
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
      <a href="/domain" class="nav-link">Domain</a>
      <a href="/api" class="nav-link" style="color:var(--text)">API</a>
    </div>
  </nav>

  <div style="max-width:900px;margin:0 auto;padding:32px 24px 0">
    <h1 style="font-size:36px;font-weight:700;margin-bottom:8px">API Documentation</h1>
    <p style="color:var(--muted);font-size:15px">REST API tích hợp TempMail. Public mode không cần xác thực; private mode hỗ trợ bearer token.</p>
  </div>

  <div class="doc-tabs" id="doc-tabs">
    <button class="doc-tab active" data-tab="intro">Giới thiệu</button>
    <button class="doc-tab" data-tab="endpoints">API Endpoints</button>
    <button class="doc-tab" data-tab="examples">Ví dụ</button>
    <button class="doc-tab" data-tab="errors">Xử lý lỗi</button>
    <button class="doc-tab" data-tab="notes">Ghi chú</button>
  </div>

  <div class="doc-wrap">

    <!-- Introduction -->
    <div class="doc-section active" id="tab-intro">
      <p class="section-intro">TempMail cung cấp REST API đơn giản để tích hợp hộp thư tạm thời vào ứng dụng của bạn. Public deployment dùng <code>AUTH_DISABLED=true</code> nên không cần API key. Nếu bật private mode, gửi header <code>Authorization: Bearer &lt;API_TOKEN&gt;</code>.</p>
      <div class="base-url-box">
        <span class="label">BASE URL</span>
        <code id="base-url-val"></code>
        <button class="copy-btn" onclick="copyText(document.getElementById('base-url-val').textContent,this)">Copy</button>
      </div>
      <div class="toc">
        <h3>📋 Mục lục</h3>
        <ol>
          <li><a href="#" onclick="switchTab('intro')">Giới thiệu</a></li>
          <li><a href="#" onclick="switchTab('endpoints')">API Endpoints</a>
            <ul class="sub">
              <li>→ <a href="#" onclick="switchTab('endpoints')">Lấy danh sách domain</a></li>
              <li>→ <a href="#" onclick="switchTab('endpoints')">Tạo địa chỉ email</a></li>
              <li>→ <a href="#" onclick="switchTab('endpoints')">Xem hộp thư</a></li>
              <li>→ <a href="#" onclick="switchTab('endpoints')">Đọc email</a></li>
              <li>→ <a href="#" onclick="switchTab('endpoints')">Lấy OTP</a></li>
              <li>→ <a href="#" onclick="switchTab('endpoints')">Xóa email / hộp thư</a></li>
            </ul>
          </li>
          <li><a href="#" onclick="switchTab('examples')">Ví dụ sử dụng</a></li>
          <li><a href="#" onclick="switchTab('errors')">Xử lý lỗi</a></li>
          <li><a href="#" onclick="switchTab('notes')">Ghi chú & Giới hạn</a></li>
        </ol>
      </div>
    </div>

    <!-- Endpoints -->
    <div class="doc-section" id="tab-endpoints">

      <div class="ep-block">
        <div class="ep-title">
          <h3>1. Lấy danh sách domain</h3>
          <div class="ep-line"><span class="method GET">GET</span><span class="ep-path">/api/v1/domains</span></div>
          <p class="ep-desc">Trả về danh sách các domain email đang hoạt động.</p>
        </div>
        <div class="ep-body">
          <div class="section-label">RESPONSE</div>
          <div class="code-block">{ "domains": ["monet.uno"] }</div>
        </div>
      </div>

      <div class="ep-block">
        <div class="ep-title">
          <h3>2. Tạo địa chỉ email tạm thời</h3>
          <div class="ep-line"><span class="method POST">POST</span><span class="ep-path">/api/v1/email/generate</span></div>
          <p class="ep-desc">Tạo một địa chỉ email tạm mới. Có thể chỉ định tên người dùng và domain tuỳ chọn.</p>
        </div>
        <div class="ep-body">
          <div class="section-label">PARAMETERS (body, tuỳ chọn)</div>
          <table class="param-table">
            <thead><tr><th>Tham số</th><th>Kiểu</th><th>Mô tả</th></tr></thead>
            <tbody>
              <tr><td><code class="param-name">domain</code></td><td class="param-type">string, optional</td><td class="param-desc">Domain cụ thể (phải nằm trong danh sách domain khả dụng). Mặc định: domain đầu tiên.</td></tr>
              <tr><td><code class="param-name">user</code></td><td class="param-type">string, optional</td><td class="param-desc">Username tùy chỉnh. Nếu không điền, hệ thống sẽ tạo ngẫu nhiên.</td></tr>
            </tbody>
          </table>
          <div class="section-label" style="margin-top:16px">RESPONSE</div>
          <div class="code-block">{ "email": "john.doe.4f2a@monet.uno", "user": "john.doe.4f2a", "domain": "monet.uno" }</div>
        </div>
      </div>

      <div class="ep-block">
        <div class="ep-title">
          <h3>3. Xem danh sách email trong hộp thư</h3>
          <div class="ep-line"><span class="method GET">GET</span><span class="ep-path">/api/v1/email/{domain}/{user}/messages</span></div>
          <p class="ep-desc">Lấy danh sách tất cả email đã nhận trong hộp thư.</p>
        </div>
        <div class="ep-body">
          <div class="section-label">PARAMETERS (path)</div>
          <table class="param-table">
            <thead><tr><th>Tham số</th><th>Kiểu</th><th>Mô tả</th></tr></thead>
            <tbody>
              <tr><td><code class="param-name">domain</code><span class="required">*</span></td><td class="param-type">string, path</td><td class="param-desc">Domain email (vd: monet.uno)</td></tr>
              <tr><td><code class="param-name">user</code><span class="required">*</span></td><td class="param-type">string, path</td><td class="param-desc">Username (phần trước @)</td></tr>
              <tr><td><code class="param-name">page</code></td><td class="param-type">number, query</td><td class="param-desc">Trang hiện tại (mặc định: 1)</td></tr>
            </tbody>
          </table>
          <div class="section-label" style="margin-top:16px">RESPONSE</div>
          <div class="code-block">{ "messages": [{ "id": "abc-123", "from": "sender@example.com", "subject": "Hello", "receivedAt": "2026-05-08T06:00:00.000Z", "otp": null, "size": 2048 }] }</div>
        </div>
      </div>

      <div class="ep-block">
        <div class="ep-title">
          <h3>4. Đọc nội dung email</h3>
          <div class="ep-line"><span class="method GET">GET</span><span class="ep-path">/api/v1/email/{domain}/{user}/messages/{id}</span></div>
          <p class="ep-desc">Lấy nội dung đầy đủ của một email theo ID.</p>
        </div>
        <div class="ep-body">
          <div class="section-label">RESPONSE</div>
          <div class="code-block">{ "id": "abc-123", "from": "sender@example.com", "to": "john.doe.4f2a@monet.uno", "subject": "Hello", "text": "Nội dung plain text...", "html": "&lt;p&gt;Nội dung HTML...&lt;/p&gt;", "receivedAt": "2026-05-08T06:00:00.000Z" }</div>
        </div>
      </div>

      <div class="ep-block">
        <div class="ep-title">
          <h3>5. Lấy mã OTP / 2FA mới nhất</h3>
          <div class="ep-line"><span class="method GET">GET</span><span class="ep-path">/api/v1/email/{domain}/{user}/otp</span></div>
          <p class="ep-desc">Tự động trích xuất mã OTP/2FA từ email mới nhất trong hộp thư.</p>
        </div>
        <div class="ep-body">
          <div class="section-label">RESPONSE</div>
          <div class="code-block">{ "otp": "123456", "receivedAt": "2026-05-08T06:00:00.000Z" }</div>
        </div>
      </div>

      <div class="ep-block">
        <div class="ep-title">
          <h3>6. Xóa một email</h3>
          <div class="ep-line"><span class="method DELETE">DELETE</span><span class="ep-path">/api/v1/email/{domain}/{user}/messages/{id}</span></div>
          <p class="ep-desc">Xóa một email cụ thể theo ID.</p>
        </div>
        <div class="ep-body">
          <div class="section-label">RESPONSE</div>
          <div class="code-block">{ "deleted": true }</div>
        </div>
      </div>

      <div class="ep-block">
        <div class="ep-title">
          <h3>7. Xóa toàn bộ hộp thư</h3>
          <div class="ep-line"><span class="method DELETE">DELETE</span><span class="ep-path">/api/v1/email/{domain}/{user}</span></div>
          <p class="ep-desc">Xóa tất cả email trong hộp thư và hủy địa chỉ email.</p>
        </div>
        <div class="ep-body">
          <div class="section-label">RESPONSE</div>
          <div class="code-block">{ "deleted": 5 }</div>
        </div>
      </div>
    </div>

    <!-- Examples -->
    <div class="doc-section" id="tab-examples">
      <p class="section-intro">Ví dụ sử dụng API public với <code style="background:var(--surface2);padding:1px 6px;border-radius:3px">curl</code>. Thay <code style="background:var(--surface2);padding:1px 6px;border-radius:3px">BASE_URL</code> bằng URL thực của bạn.</p>

      <div class="section-label">1. Tạo email ngẫu nhiên</div>
      <div class="code-block" style="position:relative"><button class="copy-btn" onclick="copyCode(this)">Copy</button>curl -s -X POST "BASE_URL/api/v1/email/generate" | jq</div>

      <div class="section-label" style="margin-top:20px">2. Tạo email với username tuỳ chỉnh</div>
      <div class="code-block" style="position:relative"><button class="copy-btn" onclick="copyCode(this)">Copy</button>curl -s -X POST "BASE_URL/api/v1/email/generate" \
  -H "Content-Type: application/json" \
  -d '{"user":"myname","domain":"monet.uno"}' | jq</div>

      <div class="section-label" style="margin-top:20px">3. Kiểm tra hộp thư</div>
      <div class="code-block" style="position:relative"><button class="copy-btn" onclick="copyCode(this)">Copy</button>curl -s "BASE_URL/api/v1/email/monet.uno/john.doe.4f2a/messages" | jq</div>

      <div class="section-label" style="margin-top:20px">4. Lấy mã OTP tự động</div>
      <div class="code-block" style="position:relative"><button class="copy-btn" onclick="copyCode(this)">Copy</button>curl -s "BASE_URL/api/v1/email/monet.uno/john.doe.4f2a/otp" | jq</div>

      <div class="section-label" style="margin-top:20px">5. Đọc nội dung email</div>
      <div class="code-block" style="position:relative"><button class="copy-btn" onclick="copyCode(this)">Copy</button>curl -s "BASE_URL/api/v1/email/monet.uno/john.doe.4f2a/messages/abc-123" | jq</div>
    </div>

    <!-- Error Handling -->
    <div class="doc-section" id="tab-errors">
      <p class="section-intro">Tất cả responses đều có cấu trúc thống nhất. Khi thành công, <code style="background:var(--surface2);padding:1px 6px;border-radius:3px">success: true</code>. Khi lỗi, <code style="background:var(--surface2);padding:1px 6px;border-radius:3px">success: false</code> kèm error object.</p>

      <div class="section-label">CẤU TRÚC RESPONSE LỖI</div>
      <div class="code-block">{ "success": false, "data": null, "error": { "code": "not_found", "message": "Mailbox not found" } }</div>

      <div class="section-label" style="margin-top:24px">MÃ LỖI</div>
      <table class="error-table">
        <thead><tr><th>HTTP Status</th><th>Error Code</th><th>Mô tả</th></tr></thead>
        <tbody>
          <tr><td>400</td><td><code class="param-name">invalid_domain</code></td><td class="param-desc">Domain không tồn tại hoặc chưa được kích hoạt</td></tr>
          <tr><td>404</td><td><code class="param-name">not_found</code></td><td class="param-desc">Hộp thư hoặc email không tìm thấy</td></tr>
          <tr><td>404</td><td><code class="param-name">otp_not_found</code></td><td class="param-desc">Không tìm thấy mã OTP trong hộp thư</td></tr>
          <tr><td>413</td><td><code class="param-name">message_too_large</code></td><td class="param-desc">Email vượt quá giới hạn kích thước cho phép</td></tr>
          <tr><td>429</td><td><code class="param-name">rate_limited</code></td><td class="param-desc">Quá nhiều request. Thử lại sau vài giây</td></tr>
          <tr><td>404</td><td><code class="param-name">route_not_found</code></td><td class="param-desc">Endpoint không tồn tại</td></tr>
        </tbody>
      </table>
    </div>

    <!-- Notes -->
    <div class="doc-section" id="tab-notes">
      <div class="note-box">⚠️ Email tạm thời sẽ tự động bị xóa sau <strong>3 ngày</strong> kể từ khi nhận.</div>
      <div class="note-box" style="background:rgba(59,130,246,.06);border-color:rgba(59,130,246,.2);color:#60a5fa">ℹ️ Rate limit: tối đa <strong>120 request / 60 giây</strong> mỗi IP. Vượt quá sẽ nhận lỗi 429.</div>
      <div class="note-box" style="background:rgba(34,197,94,.06);border-color:rgba(34,197,94,.2);color:#4ade80">✅ Public mode đang bật để mọi người dùng UI và API không cần token. Rate limit vẫn bảo vệ chống lạm dụng cơ bản.</div>
      <p style="color:var(--muted);font-size:13px;margin-top:20px;line-height:1.7">
        Hệ thống chạy trên Cloudflare Workers + D1 + R2 — đảm bảo tốc độ cao và độ trễ thấp trên toàn cầu.
        Email được xử lý qua Cloudflare Email Routing và lưu trữ tạm trong 3 ngày trước khi tự động dọn dẹp.
      </p>
    </div>

  </div>

  <script>
    document.getElementById('base-url-val').textContent = location.origin;
    function switchTab(name){
      document.querySelectorAll('.doc-tab').forEach(t=>t.classList.toggle('active',t.dataset.tab===name));
      document.querySelectorAll('.doc-section').forEach(s=>s.classList.toggle('active',s.id==='tab-'+name));
    }
    document.querySelectorAll('.doc-tab').forEach(t=>t.onclick=()=>switchTab(t.dataset.tab));
    function copyText(text,btn){
      navigator.clipboard.writeText(text).then(()=>{btn.textContent='Copied!';setTimeout(()=>btn.textContent='Copy',1500)});
    }
    function copyCode(btn){
      const text=btn.parentElement.textContent.replace('Copy','').trim();
      copyText(text,btn);
    }
  </script>
</body>
</html>`;
