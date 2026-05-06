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
    "imap_tls": true,
    "default_domain": "yourdomain.com"
  }
  ```

---

## 2. 📱 ViOTP SMS (tùy chọn — chỉ cần khi gặp phone challenge)

- [ ] Đăng ký tài khoản tại https://viotp.vn
- [ ] Nạp tiền (50k–100k VNĐ đủ để test)
- [ ] Lấy **API Token** từ dashboard ViOTP
- [ ] Tìm **Service ID** của OpenAI trên ViOTP
  - Vào mục "Dịch vụ" → tìm "OpenAI" hoặc "ChatGPT"
- [ ] Điền vào `config.json`:
  ```json
  {
    "viotp_token": "your-api-token-here",
    "viotp_service_id": 5
  }
  ```

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
- [ ] Test ViOTP balance: chạy `--viotp-token xxx` và xem log balance check
- [ ] Chạy thử 1 account: `./chatgpt-creator register --total 1 --pacing none`

---

## 5. 🚀 Lệnh chạy mẫu

```bash
# Chạy đầy đủ với tất cả tính năng
./chatgpt-creator register \
  --total 10 \
  --workers 2 \
  --pacing human \
  --proxy-list proxies.txt \
  --viotp-token "your-token" \
  --output results.txt

# Chạy nhanh (test, không pacing, không proxy)
./chatgpt-creator register --total 1 --pacing none
```

---

## 6. 📝 Ghi chú

- **Pacing `human`** (120–300s giữa các lần): dùng cho production, IP tĩnh
- **Pacing `fast`** (5–15s): dùng khi có rotating residential proxy
- **Pacing `none`**: chỉ dùng khi test, dễ bị 429
- Kết quả lưu tại `results.txt` theo format `email|password`
- IMAP password phải là **App Password**, không phải password Gmail thường
