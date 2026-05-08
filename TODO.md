# TODO — Chuẩn bị trước khi chạy chatgpt-creator

## 1. 📧 Mail Domain + IMAP (bắt buộc để ổn định)

- [ ] Mua domain rẻ (`.xyz`, `.shop`, `.online` — khoảng $1–5/năm)
  - Gợi ý: Namecheap, Cloudflare Registrar, Porkbun
- [ ] Trỏ domain về Cloudflare (đổi nameserver)
- [ ] Vào Cloudflare → **Email Routing** → bật Catch-all
  - Action: Forward to → điền email Gmail của bạn
- [ ] Bật **Gmail IMAP**
  - Gmail Settings → See all settings → Forwarding and POP/IMAP → Enable IMAP
- [ ] Tạo **Gmail App Password** (bắt buộc nếu bật 2FA)
  - myaccount.google.com → Security → App Passwords → Mail
- [ ] Điền vào `config.json`:
  ```json
  {
    "imap_host": "imap.gmail.com",
    "imap_port": 993,
    "imap_user": "you@gmail.com",
    "imap_password": "xxxx xxxx xxxx xxxx",
    "imap_use_tls": true,
    "default_domain": "yourdomain.com"
  }
  ```

---

## 2. 📱 ViOTP SMS (không hỗ trợ trong safe mode)

- [ ] Không cấu hình `viotp_token` hoặc `viotp_service_id` cho runtime hiện tại.
- [ ] Nếu bị phone challenge, flow sẽ trả về `phone_challenge` để xử lý thủ công.
- [ ] `docs/viotp-api.md` chỉ là tài liệu tham khảo API, không phải feature đang bật.

---

## 3. 🌐 Proxy List (tùy chọn — nên có để tránh bị block IP)

- [ ] Chuẩn bị danh sách proxy (residential hoặc datacenter)
  - Format: 1 proxy/dòng, ví dụ: `http://user:pass@ip:port`
  - Comment bằng `#`
- [ ] Lưu vào file, ví dụ: `proxies.txt`
- [ ] Dùng flag `--proxy-list proxies.txt` khi chạy

---

## 4. ✅ Kiểm tra trước khi chạy batch

- [ ] `go build ./...` — đảm bảo build thành công
- [ ] `go test ./...` — đảm bảo tests pass
- [ ] Test IMAP thủ công: gửi email thử tới `test@yourdomain.com`, kiểm tra Gmail nhận được
- [ ] Xác nhận không truyền `--viotp-token`/`--viotp-service-id` (runtime sẽ fail-closed nếu truyền)
- [ ] Chạy thử 1 account: `./chatgpt-creator register --total 1 --pacing none`

---

## 5. 🚀 Lệnh chạy mẫu

```bash
# Chạy production-safe: proxy pool + IMAP catch-all
./chatgpt-creator register \
  --total 10 \
  --workers 2 \
  --pacing human \
  --proxy-list proxies.txt \
  --output results.txt

# Chạy nhanh (test, không pacing, không proxy)
./chatgpt-creator register --total 1 --pacing none
```

---

## 6. Risks / concerns

- [ ] Coverage tổng hiện khoảng `46.2%`, dưới mục tiêu 80%.
- [ ] `cmd/register/command.go` đã lớn; cân nhắc tách nhỏ khi chỉnh CLI tiếp theo.
- [ ] IMAP parser hiện tối giản; kiểm tra thêm với Gmail thực tế và server trả literal/multiline response.
- [ ] ViOTP/Codex flags vẫn hiện trong CLI nhưng safe mode chặn fail-closed; cân nhắc làm UX rõ hơn.
- [ ] Interactive mode chưa fail exit code khi batch không đạt target; non-interactive đã xử lý.

---

## 7. 📝 Ghi chú

- **Pacing `human`** (120–300s giữa các lần): dùng cho production, IP tĩnh
- **Pacing `fast`** (5–15s): dùng khi có rotating residential proxy
- **Pacing `none`**: chỉ dùng khi test, dễ bị 429
- ViOTP (`--viotp-token`, `--viotp-service-id`) và Codex (`--codex`) bị chặn fail-closed trong safe mode
- Phone challenge hiện là detection-only (`phone_challenge`), không tự động thuê số/SMS OTP
- Kết quả lưu tại `results.txt` theo format `email|password`
- IMAP password phải là **App Password**, không phải password Gmail thường
