# ChatGPT TOTP 2FA Setup — API Contract

Discovered via browser interception on 2026-05-09. All endpoints under `chatgpt.com`.

## Constraints

- `chatgpt.com` has Cloudflare protection — raw `curl` is blocked. Must use browser session (camofox/chromedp) with valid cookies.
- `monet.uno` is a Google Workspace domain → OpenAI's unified login page forces Google SSO. Must bypass via the signin API flow below.
- `activate_enrollment` requires a valid `Authorization: Bearer` access token.

---

## Step 1 — Login (bypass Google SSO for Google Workspace domains)

### 1a. Get CSRF token

```
GET /api/auth/csrf
→ { "csrfToken": "..." }
```

### 1b. Get authorize URL

```
POST /api/auth/signin/openai?prompt=login&login_hint={email}
Content-Type: application/x-www-form-urlencoded
Body: callbackUrl=https%3A%2F%2Fchatgpt.com%2F&csrfToken={csrf}&json=true

→ { "url": "https://auth.openai.com/api/accounts/authorize?..." }
```

### 1c. Navigate browser to authorize URL

Navigate to the `url` from 1b. OpenAI sends an email OTP (magic link code) to the email address.

Response page: `auth.openai.com` shows "Check your inbox" form.

### 1d. Fetch OTP from mail API

```
GET https://mail.monet.uno/api/v1/email/{domain}/{user}/otp
User-Agent: Mozilla/5.0 ...   ← required, Cloudflare blocks without it

→ { "success": true, "data": { "otp": "123456", "receivedAt": "2026-05-09T07:45:23.621Z" } }
```

Poll until `status == "received"`. Timeout: 300s.

### 1e. Submit OTP in browser form

Type OTP into the code field and click Continue. Browser redirects to `chatgpt.com` with a valid session.

### 1f. Get access token

```js
// JS in browser context
fetch('/api/auth/session').then(r => r.json())
→ { "accessToken": "eyJ...", "user": { "email": "...", "mfa": false } }
```

Access token JWT claims relevant to MFA:
- `pwd_auth_time` — timestamp of last auth (set even for OTP login)
- `https://api.openai.com/auth.chatgpt_account_id` — account UUID needed for other operations

---

## Step 2 — Enroll TOTP Factor

```
POST /backend-api/accounts/mfa/enroll
Authorization: Bearer {access_token}
Content-Type: application/json

{ "factor_type": "totp" }
```

Response `200 OK`:

```json
{
  "secret": "B6P2A2GVK4ERXUWUZDYBJHQMFHAFWA3V",
  "email": null,
  "session_id": "69fee683e2748191a9b08f26e67d4ee0",
  "factor": {
    "id": "69fee68420c08191b1d8fdca5a7060d8",
    "factor_type": "totp",
    "is_recovery": false,
    "metadata": null
  }
}
```

The `secret` is a base32-encoded TOTP secret. The QR code URI format is:
```
otpauth://totp/OpenAI%3A{email}?secret={secret}&issuer=OpenAI
```

Valid `factor_type` values: `totp`, `recovery_code`, `push_auth`, `sms`, `passkey`

---

## Step 3 — Generate TOTP Code

Pure stdlib (no external library):

```go
func generateTOTP(secret string) (string, error) {
    key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(
        strings.ToUpper(strings.ReplaceAll(secret, " ", "")),
    )
    if err != nil {
        return "", err
    }
    counter := uint64(time.Now().Unix()) / 30
    msg := make([]byte, 8)
    binary.BigEndian.PutUint64(msg, counter)
    mac := hmac.New(sha1.New, key)
    mac.Write(msg)
    h := mac.Sum(nil)
    offset := h[len(h)-1] & 0xf
    code := binary.BigEndian.Uint32(h[offset:offset+4]) & 0x7fffffff
    return fmt.Sprintf("%06d", code%1_000_000), nil
}
```

```python
# Python equivalent (no pyotp needed)
import hmac, hashlib, struct, time, base64

def generate_totp(secret: str) -> str:
    key = base64.b32decode(secret.upper().replace(" ", ""))
    counter = int(time.time()) // 30
    msg = struct.pack(">Q", counter)
    h = hmac.new(key, msg, hashlib.sha1).digest()
    offset = h[-1] & 0xf
    code = struct.unpack(">I", h[offset:offset+4])[0] & 0x7fffffff
    return str(code % 1_000_000).zfill(6)
```

---

## Step 4 — Activate Enrollment

```
POST /backend-api/accounts/mfa/user/activate_enrollment
Authorization: Bearer {access_token}
Content-Type: application/json

{
  "session_id":  "{session_id from Step 2}",
  "factor_id":   "{factor.id from Step 2}",
  "factor_type": "totp",
  "code":        "{6-digit TOTP}"
}
```

Response `200 OK` → MFA enabled.

Error responses:
- `{ "error": { "message": "Invalid code", "code": "invalid_code" } }` — wrong TOTP, retry with fresh code
- `{ "detail": [{ "type": "missing", "loc": ["body", "factor_type"] }] }` — missing `factor_type` field

---

## Step 5 — Verify MFA Status

```
GET /backend-api/accounts/mfa_info
Authorization: Bearer {access_token}

→ {
    "mfa_enabled": true,
    "mfa_enabled_v2": true,
    "native_default_factor_id": "...",
    "factors": {
      "totp": [{ "id": "...", "factor_type": "totp", "is_recovery": false }],
      "push_auth": null,
      "passkeys": [],
      "sms": []
    }
  }
```

---

## Other Relevant Endpoints (discovered)

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/backend-api/accounts/mfa_info` | GET | Get MFA status and enrolled factors |
| `/backend-api/accounts/mfa_push_auth_devices` | GET | List push auth devices |
| `/backend-api/accounts/security_settings/info` | GET | Security settings overview |
| `/backend-api/accounts/mfa/enroll` | POST | Start factor enrollment |
| `/backend-api/accounts/mfa/user/activate_enrollment` | POST | Confirm enrollment with code |
| `/backend-api/accounts/sessions` | GET | List active sessions |

---

## UI Flow Notes (for reference)

When enabling MFA via the Settings > Security UI, the flow differs slightly:
1. Clicking the "Authenticator app" toggle redirects to `auth.openai.com/email-verification`
2. User must verify identity with an email OTP first
3. After OTP, redirected back to `chatgpt.com/?action=enable&factor=totp#settings/Security`
4. UI calls `/backend-api/accounts/mfa/enroll` → shows QR code
5. User enters TOTP → UI calls `/backend-api/accounts/mfa/user/activate_enrollment`

The programmatic flow (Steps 1–4 above) skips the UI email-verification step.
