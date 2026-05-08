# Proxy.vn API xoay

## Mục tiêu

Ghi lại API proxy xoay từ `https://proxy.vn/?home=apixoay` và cách dùng an toàn với luồng tạo tài khoản hiện có.

## Trạng thái app hiện tại

App đã có cơ chế chạy đa luồng ở `internal/register/batch.go`:

- `maxWorkers` tạo nhiều goroutine chạy song song.
- `sync.WaitGroup` chờ worker hoàn tất.
- `sync/atomic` quản lý counters và stop flag.
- Mutex bảo vệ log và ghi file.
- `ProviderOptions.ProxyPool` cho phép dùng proxy pool thay vì một proxy cố định.

App cũng đã có proxy pool ở `internal/proxy/pool.go`:

- `RoundRobinPool` xoay proxy theo vòng.
- Proxy lỗi bị cooldown.
- Có thống kê success, failures, health.
- Có loader đọc danh sách proxy từ file, mỗi dòng một proxy URL.

## API lấy proxy xoay

Endpoint:

```text
https://proxyxoay.shop/api/get.php
```

Hỗ trợ `GET` hoặc `POST`.

Tham số từ trang proxy.vn:

| Tham số | Ý nghĩa |
| --- | --- |
| `key` | Key xoay nhận được khi mua hàng |
| `nhamang` | Nhà mạng, theo danh sách proxy.vn cung cấp |
| `tinhthanh` | Mã tỉnh/thành, theo danh sách proxy.vn cung cấp |
| `whitelist` | IPv4 được phép sử dụng proxy |

Response thành công:

```json
{
  "status": 100,
  "message": "proxy nay se die sau 1777s",
  "proxyhttp": "42.117.243.215:10836::",
  "proxysocks5": "42.117.243.215:30836::",
  "Nha Mang": "fpt",
  "Vi Tri": "HaNoi1",
  "Token expiration date": "22:52 19-02-2025"
}
```

Lỗi được tài liệu trang gốc nêu: `status=101`, `status=102`.

## API mua key xoay

Endpoint theo thời hạn:

```text
https://proxy.vn/proxyxoay/apimuangay.php
https://proxy.vn/proxyxoay/apimuatuan.php
https://proxy.vn/proxyxoay/apimuathang.php
```

Tham số:

| Tham số | Ý nghĩa |
| --- | --- |
| `key` | Key tài khoản proxy.vn |
| `thoigian` | Số đơn vị thời gian: ngày, tuần hoặc tháng theo endpoint |
| `soluong` | Số lượng key xoay cần mua |

Response thành công có thể trả về:

```json
{
  "status": 100,
  "keyxoay": "rwywzSOvFNZOWDVJJBrQRb"
}
```

hoặc:

```json
{
  "status": 100,
  "soluong": 1,
  "comen": "successful transaction 1 key xoay"
}
```

Lỗi ví dụ:

```json
{
  "status": 101,
  "comen": "key does not exist"
}
```

## API gia hạn key xoay

Endpoint theo thời hạn:

```text
https://proxy.vn/proxyxoay/apigiahanngay.php
https://proxy.vn/proxyxoay/apigiahantuan.php
https://proxy.vn/proxyxoay/apigiahanthang.php
```

Tham số:

| Tham số | Ý nghĩa |
| --- | --- |
| `key` | Key tài khoản proxy.vn |
| `keyxoay` | Key xoay cần gia hạn |
| `thoigian` | Số đơn vị thời gian: ngày, tuần hoặc tháng theo endpoint |

Response thành công:

```json
{
  "status": 100,
  "comen": "Renewal successful"
}
```

Lỗi ví dụ:

```json
{
  "status": 101,
  "comen": "Not determined"
}
```

## API lấy danh sách key xoay còn hạn

Endpoint:

```text
https://proxy.vn/proxyxoay/apigetkeyxoay.php
```

Tham số:

| Tham số | Ý nghĩa |
| --- | --- |
| `key` | Key tài khoản proxy.vn |

Response thành công có thể gồm nhiều record:

```json
{
  "status": 100,
  "keyxoay": "oJKZjfhjfjh4sXVWqLnbs",
  "expired": "21:43 29-03-25"
}
```

Lỗi ví dụ:

```json
{
  "status": 101,
  "comen": "key does not exist"
}
```

## Cách dùng với app hiện tại

Cách đơn giản nhất hiện tại:

1. Gọi API lấy proxy xoay bằng key riêng.
2. Chọn `proxyhttp` hoặc `proxysocks5`.
3. Chuẩn hóa thành proxy URL mà Go HTTP client hỗ trợ, ví dụ:

```text
http://42.117.243.215:10836
socks5://42.117.243.215:30836
```

4. Lưu vào file proxy local, mỗi dòng một proxy.
5. Trỏ CLI/web config tới file proxy list hiện có.

Không commit file proxy local vì có thể chứa proxy/IP/key nhạy cảm.

## Kế hoạch tích hợp API xoay động

### Phase 1: Provider đọc proxy.vn

- Thêm provider nhỏ trong `internal/proxy` gọi `get.php` bằng key xoay.
- Input cấu hình: API key, nhà mạng, tỉnh thành, whitelist, loại proxy ưu tiên.
- Output: proxy URL chuẩn cho batch runner.
- Không log key. Redact URL nếu chứa credential.

### Phase 2: Cache theo TTL

- Parse TTL từ `message`, ví dụ `proxy nay se die sau 1777s`.
- Cache proxy cho tới gần hết TTL.
- Refresh trước khi proxy chết.
- Nếu response lỗi `101/102`, trả lỗi typed để batch dừng hoặc backoff.

### Phase 3: Gắn vào worker pool

- Implement `ProxyPool.Next(ctx)` gọi provider khi cần proxy mới.
- `Report(proxyURL, success)` dùng failure count để quyết định refresh sớm.
- `Stats()` giữ health metrics hiện có.

### Phase 4: CLI/web config

- Thêm flags/config:
  - `proxy-vn-key`
  - `proxy-vn-network`
  - `proxy-vn-province`
  - `proxy-vn-whitelist`
  - `proxy-vn-type=http|socks5`
- Ưu tiên config: defaults < file < env < flags.
- Env dùng cho secrets, không lưu plaintext key trong repo.

### Phase 5: Test và safety

- Unit test parser response thành công/lỗi.
- Unit test TTL extraction.
- Unit test không log key.
- Race test cho pool dưới nhiều worker.
- Integration test dùng fake HTTP server, không gọi proxy.vn thật.

## Khuyến nghị

Nên giữ bước đầu là file proxy list local nếu cần chạy ngay. Tích hợp API proxy.vn động chỉ nên làm khi cần refresh proxy tự động theo TTL, vì thêm HTTP dependency và secret handling vào batch runner.
