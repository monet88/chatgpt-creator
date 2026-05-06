# ViOTP API Documentation

> ⚠️ **Runtime status (safe mode): unsupported**
> - Tài liệu này chỉ để tham khảo API ViOTP.
> - Batch runtime hiện tại fail-closed khi nhận `viotp_token` hoặc `viotp_service_id`.
> - Phone challenge hiện là detection-only (`phone_challenge`), không tự động thuê số/nhận OTP qua ViOTP.

> **Lưu ý !**
> - Sử dụng giao thức GET cho mọi truy vấn.
> - Dữ liệu trả về dưới dạng JSON.
> - Số điện thoại chỉ bao gồm 9 chữ số, không có số 0 ở đầu.
> - Dịch vụ Zalo hiện tại không hỗ trợ thuê số qua API

---

## 0. Tra cứu thông tin tài khoản

### 0.1 Ví dụ gửi đi
```http
GET https://api.viotp.com/users/balance?token=5abec70115c70ebb685169fe7dd985e7
```
**Danh sách tham số gửi đi:**
- `token` : API token của bạn. Dùng để định danh người gửi

### 0.2 Kết quả trả về
```json
{
  "status_code": 200,
  "success": true,
  "message": "successful",
  "data": {
    "balance": 9999999999
  }
}
```
**Danh sách trả về:**
- `status_code` : Mã của kết quả
  - `200` : Thành công
  - `401` : Lỗi xác thực
  - `-1` : Có lỗi
- `message` : Thông tin thêm (Thông tin báo thành công hoặc là nội dung của lỗi)
- `data` : Thông tin số tiền trong tài khoản
  - `balance` : Số tiền còn lại

---

## 1. Lấy danh sách nhà mạng

> Mobi, Vina, Viettel, VNMB, ITELECOM - chỉ áp dụng cho thuê SIM Việt Nam
> UNITEL, ETL, BEELINE, LAOTEL - chỉ áp dụng cho thuê SIM Lào

### 1.1 Ví dụ gửi đi
```http
GET https://api.viotp.com/networks/get?token=5abec70115c70ebb685169fe7dd985e7
```
**Danh sách tham số gửi đi:**
- `token` : API token của bạn

### 1.2 Kết quả trả về
```json
{
  "status_code": 200,
  "success": true,
  "message": "successful",
  "data": [
    {"id": 1, "name": "MOBIFONE"},
    {"id": 2, "name": "VINAPHONE"},
    {"id": 3, "name": "VIETTEL"},
    {"id": 4, "name": "VIETNAMOBILE"},
    {"id": 5, "name": "ITELECOM"},
    {"id": 6, "name": "VODAFONE"},
    {"id": 7, "name": "WINTEL"},
    {"id": 8, "name": "METFONE"},
    {"id": 9, "name": "UNITEL"},
    {"id": 10, "name": "ETL"},
    {"id": 11, "name": "BEELINE"},
    {"id": 12, "name": "LAOTEL"}
  ]
}
```
**Danh sách trả về:**
- `status_code` : Mã của kết quả (200, 401, -1)
- `message` : Thông tin thêm
- `data` : Danh sách nhà mạng
  - `id` : Id của nhà mạng
  - `name` : Tên của nhà mạng

---

## 2. Lấy danh sách dịch vụ

### 2.1 Ví dụ gửi đi
```http
GET https://api.viotp.com/service/getv2?token=5abec70115c70ebb685169fe7dd985e7&country=vn
```
**Danh sách tham số gửi đi:**
- `token` : API token của bạn
- `country` : Mã quốc gia cần lấy danh sách dịch vụ hỗ trợ (mặc định không truyền parameter `country` là lấy danh sách dịch vụ SIM Việt Nam)
  - `la`: Lào
  - `vn`: Việt Nam

### 2.2 Kết quả trả về
```json
{
  "status_code": 200,
  "success": true,
  "message": "successful",
  "data": [
    {"id": 1, "name": "Facebook", "price": 800},
    {"id": 2, "name": "Shopee", "price": 600}
  ]
}
```
**Danh sách trả về:**
- `status_code` : Mã của kết quả (200, 401, -1)
- `message` : Thông tin thêm
- `data` : Danh sách dịch vụ
  - `id` : Id của dịch vụ
  - `name` : Tên của dịch vụ
  - `price` : Giá thuê dịch vụ

---

## 3. Yêu cầu dịch vụ

### 3.1 Ví dụ gửi đi

**THUÊ SIM VIỆT NAM**
- Cơ bản:
```http
GET https://api.viotp.com/request/getv2?token=5abec70115c70ebb685169fe7dd985e7&serviceId=1
```
- Tùy chọn nhà mạng:
```http
GET https://api.viotp.com/request/getv2?token=5abec70115c70ebb685169fe7dd985e7&serviceId=1&network=MOBIFONE|VINAPHONE|VIETTEL|VIETNAMOBILE|ITELECOM|WINTEL
```
- Tùy chọn đầu số muốn lấy:
```http
GET https://api.viotp.com/request/getv2?token=5abec70115c70ebb685169fe7dd985e7&serviceId=1&prefix=90|91|92|93
```
- Tùy chọn đầu số không muốn lấy:
```http
GET https://api.viotp.com/request/getv2?token=5abec70115c70ebb685169fe7dd985e7&serviceId=1&exceptPrefix=94|96|97|98
```
- Tùy chọn thuê lại số cũ:
```http
GET https://api.viotp.com/request/getv2?token=5abec70115c70ebb685169fe7dd985e7&serviceId=1&number=0987654321
```

**THUÊ SIM LÀO**
- Cơ bản:
```http
GET https://api.viotp.com/request/getv2?token=5abec70115c70ebb685169fe7dd985e7&serviceId=1&country=la
```
- Tùy chọn nhà mạng:
```http
GET https://api.viotp.com/request/getv2?token=5abec70115c70ebb685169fe7dd985e7&serviceId=1&country=la&network=UNITEL|ETL|BEELINE|LAOTEL
```
- Tùy chọn thuê lại số cũ:
```http
GET https://api.viotp.com/request/getv2?token=5abec70115c70ebb685169fe7dd985e7&serviceId=1&number=8562098765432
```

**Danh sách tham số gửi đi:**
- `token` : API token của bạn
- `serviceId` : Id của dịch vụ (Lấy từ API 2)
- `network` : Nhà mạng muốn lấy số (có thể ngăn cách bởi `|`)
- `prefix` : Đầu số muốn lấy số
- `exceptPrefix` : Đầu số không muốn lấy số
- `number` : Số điện thoại muốn thuê lại (dùng giá trị `re_phone_number` hoặc `PhoneOriginal`)

### 3.2 Kết quả trả về
```json
{
  "status_code": 200,
  "success": true,
  "message": "Tạo yêu cầu thành công !",
  "data": {
    "phone_number": "987654321",
    "re_phone_number": "84987654321",
    "countryISO": "VN",
    "countryCode": "84",
    "balance": 50000,
    "request_id": "122314"
  }
}
```
**Danh sách trả về:**
- `status_code` : Mã của kết quả
  - `200` : Thành công
  - `401` : Lỗi xác thực
  - `429` : Limit exceeded (vượt giới hạn số chờ)
  - `-1` : Có lỗi
  - `-2` : Số dư tài khoản không đủ
  - `-3` : Kho số đang tạm hết
  - `-4` : Ứng dụng không tồn tại hoặc tạm ngưng
- `data` : Thông tin sim số
  - `phone_number` : Số điện thoại đã thuê được
  - `re_phone_number` : Giá trị dùng để thuê lại
  - `countryISO` : Mã quốc gia SIM (VN, LA)
  - `countryCode` : Mã vùng (84, 856)
  - `balance` : Số dư hiện tại
  - `request_id` : Mã để lấy Code OTP (dùng ở mục 4)

---

## 4. Lấy code của 1 số điện thoại đã lấy

### 4.1 Ví dụ gửi đi
```http
GET https://api.viotp.com/session/getv2?requestId=1234122&token=5abec70115c70ebb685169fe7dd985e7
```
**Danh sách tham số gửi đi:**
- `token` : API token của bạn
- `requestId` : Mã phiên thuê số cần lấy code (lấy ở mục 3)

### 4.2 Kết quả trả về

**Nhận được tin nhắn:**
```json
{
  "status_code": 200,
  "success": true,
  "message": "successful",
  "data": {
    "ID": 58098,
    "ServiceID": 1,
    "ServiceName": "Momo",
    "Status": 1,
    "Price": 350,
    "Phone": "987654321",
    "SmsContent": "486460 la ma xac thuc OTP dang ky vi MoMo...",
    "IsSound": false,
    "CreatedTime": "2020-08-06T17:13:24.88",
    "Code": "486460",
    "PhoneOriginal": "0987654321",
    "CountryISO": "VN",
    "CountryCode": "84"
  }
}
```

**Nhận được cuộc gọi:**
```json
{
  "status_code": 200,
  "success": true,
  "message": "successful",
  "data": {
    "ID": 58060,
    "ServiceID": 1,
    "ServiceName": "Momo",
    "Status": 1,
    "Price": 350,
    "Phone": "987654321",
    "SmsContent": "https://audio.viotp.com/12345.wav",
    "IsSound": true,
    "CreatedTime": "2020-08-06T17:13:24.88",
    "Code": "486460",
    "PhoneOriginal": "0987654321",
    "CountryISO": "VN",
    "CountryCode": "84"
  }
}
```
**Danh sách trả về:**
- `status_code` : Mã của kết quả (200, 401, 429, -1, -2 - Mã phiên không đúng)
- `data` : 
  - `Status` : Trạng thái tin nhắn (`1`: Hoàn thành, `0`: Đợi tin nhắn, `2`: Hết hạn)
  - `IsSound` : OTP là audio (true/false)
  - `SmsContent` : Nội dung tin nhắn hoặc link file audio
  - `Code` : Mã của OTP

---

## 5. Lịch sử thuê số

### 5.1 Ví dụ gửi đi
```http
GET https://api.viotp.com/session/historyv2?token=5abec70115c70ebb685169fe7dd985e7&service=1&status=2&limit=100&fromDate=2020-11-12&toDate=2020-11-12
```
**Danh sách tham số gửi đi:**
- `token` : API token của bạn
- `service` : Id dịch vụ
- `status` : Trạng thái OTP (1, 0, 2)
- `limit` : Số dữ liệu cần lấy
- `fromDate` / `toDate` : Từ ngày - đến ngày (yyyy-MM-dd)

### 5.2 Kết quả trả về
```json
{
  "status_code": 200,
  "success": true,
  "message": "successful",
  "data": [
    {
      "ID": 58098,
      "ServiceID": 1,
      "ServiceName": "Momo",
      "Status": 1,
      "Price": 600,
      "Phone": "987654321",
      "SmsContent": "486460 la ma xac thuc...",
      "IsSound": false,
      "CreatedTime": "2020-08-06T17:13:24.88",
      "Code": "486460",
      "PhoneOriginal": "0987654321",
      "CountryISO": "VN",
      "CountryCode": "84"
    }
  ]
}
```
