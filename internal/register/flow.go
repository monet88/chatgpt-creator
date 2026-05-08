package register

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/monet88/chatgpt-creator/internal/codex"
	"github.com/monet88/chatgpt-creator/internal/sentinel"
	"github.com/monet88/chatgpt-creator/internal/util"
)

const (
	// Phone verification endpoints. Update if OpenAI changes the phone API.
	phoneVerifyEndpoint = authURL + "/api/accounts/phone/send"
	phoneOTPEndpoint    = authURL + "/api/accounts/phone/validate"
	phoneOTPTimeout     = 120 * time.Second
	// Codex SSO callback server timeout.
	codexCallbackTimeout = 60 * time.Second
)

// visitHomepage visits chatgpt.com to initialize session
func (c *Client) visitHomepage() error {
	var resp *http.Response
	var err error
	for retry := 0; retry < 3; retry++ {
		req, _ := http.NewRequest("GET", baseURL+"/", nil)
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
		req.Header.Set("Upgrade-Insecure-Requests", "1")

		resp, err = c.do(req)
		if err != nil {
			return err
		}

		c.log(fmt.Sprintf("Visit Homepage (Try %d)", retry+1), resp.StatusCode)

		if resp.StatusCode == 200 || resp.StatusCode == 302 || resp.StatusCode == 307 {
			resp.Body.Close()
			return nil
		}
		resp.Body.Close()
		backoff := time.Duration(500+rand.Intn(1500)) * time.Millisecond
		time.Sleep(backoff)
	}
	return fmt.Errorf("failed to visit homepage after 3 retries (status: %d)", resp.StatusCode)
}

// getCSRF retrieves the CSRF token from chatgpt.com
func (c *Client) getCSRF() (string, error) {
	req, _ := http.NewRequest("GET", baseURL+"/api/auth/csrf", nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", baseURL+"/")

	resp, err := c.do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data struct {
		CSRFToken string `json:"csrfToken"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}

	c.log("Get CSRF", resp.StatusCode)
	if data.CSRFToken == "" {
		return "", fmt.Errorf("csrf token not found")
	}
	return data.CSRFToken, nil
}

// signin initiates the signin process and returns the authorize URL
func (c *Client) signin(email, csrf string) (string, error) {
	signinURL := baseURL + "/api/auth/signin/openai"
	params := url.Values{}
	params.Set("prompt", "login")
	params.Set("ext-oai-did", c.deviceID)
	params.Set("auth_session_logging_id", util.GenerateUUID()) // Assuming util has this or use google/uuid
	params.Set("screen_hint", "login_or_signup")
	params.Set("login_hint", email)

	fullURL := signinURL + "?" + params.Encode()

	formData := url.Values{}
	formData.Set("callbackUrl", baseURL+"/")
	formData.Set("csrfToken", csrf)
	formData.Set("json", "true")

	req, _ := http.NewRequest("POST", fullURL, strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", baseURL+"/")
	req.Header.Set("Origin", baseURL)

	resp, err := c.do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}

	c.log("Signin", resp.StatusCode)
	if data.URL == "" {
		return "", fmt.Errorf("authorize url not found")
	}
	return data.URL, nil
}

// authorize visits the authorize URL and returns the final redirect URL
func (c *Client) authorize(authURL string) (string, error) {
	req, _ := http.NewRequest("GET", authURL, nil)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Referer", baseURL+"/")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := c.do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	finalURL := resp.Request.URL.String()
	c.log("Authorize", resp.StatusCode)
	return finalURL, nil
}

// register registers the user with email and password
func (c *Client) register(email, password string) (int, map[string]interface{}, error) {
	regURL := authURL + "/api/accounts/user/register"
	payload := map[string]string{
		"username": email,
		"password": password,
	}
	jsonPayload, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", regURL, bytes.NewReader(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", authURL+"/create-account/password")
	req.Header.Set("Origin", authURL)

	applyTraceHeaders(req)

	resp, err := c.do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data map[string]interface{}
	json.Unmarshal(body, &data)

	c.log("Register", resp.StatusCode)
	return resp.StatusCode, data, nil
}

// sendOTP sends the OTP to the user's email
func (c *Client) sendOTP() (int, map[string]interface{}, error) {
	otpURL := authURL + "/api/accounts/email-otp/send"
	req, _ := http.NewRequest("GET", otpURL, nil)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Referer", authURL+"/create-account/password")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := c.do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		data = map[string]interface{}{"text": string(body)}
	}

	c.log("Send OTP", resp.StatusCode)
	return resp.StatusCode, data, nil
}

// validateOTP validates the OTP code
func (c *Client) validateOTP(code string) (int, map[string]interface{}, error) {
	valURL := authURL + "/api/accounts/email-otp/validate"
	payload := map[string]string{"code": code}
	jsonPayload, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", valURL, bytes.NewReader(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", authURL+"/email-verification")
	req.Header.Set("Origin", authURL)

	applyTraceHeaders(req)

	resp, err := c.do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data map[string]interface{}
	json.Unmarshal(body, &data)

	c.log(fmt.Sprintf("Validate OTP [%s]", code), resp.StatusCode)
	return resp.StatusCode, data, nil
}

// createAccount creates the user account with name and birthdate
func (c *Client) createAccount(name, birthdate string) (int, map[string]interface{}, error) {
	createURL := authURL + "/api/accounts/create_account"
	payload := map[string]string{
		"name":      name,
		"birthdate": birthdate,
	}
	jsonPayload, _ := json.Marshal(payload)

	sentinelCreateAccount, err := sentinel.BuildSentinelToken(c.session, c.deviceID, "create_account", c.ua, c.secChUA, c.impersonate)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get sentinel auth: %v", err)
	}

	req, _ := http.NewRequest("POST", createURL, bytes.NewReader(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", authURL+"/about-you")
	req.Header.Set("Origin", authURL)
	req.Header.Set("openai-sentinel-token", sentinelCreateAccount)

	applyTraceHeaders(req)

	resp, err := c.do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data map[string]interface{}
	json.Unmarshal(body, &data)

	c.log("Create Account", resp.StatusCode)
	return resp.StatusCode, data, nil
}

// callback handles the callback URL
func (c *Client) callback(cbURL string) (int, map[string]interface{}, error) {
	if cbURL == "" {
		return 0, nil, fmt.Errorf("empty callback url")
	}

	req, _ := http.NewRequest("GET", cbURL, nil)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := c.do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	c.log("Callback", resp.StatusCode)
	return resp.StatusCode, map[string]interface{}{"final_url": resp.Request.URL.String()}, nil
}

func (c *Client) RunRegister(emailAddr, password, name, birthdate string) error {
	return c.RunRegisterWithContext(context.Background(), emailAddr, password, name, birthdate)
}

// RunRegisterWithContext executes the full registration flow and, if codex extraction
// is enabled, extracts OAuth tokens for Codex after a successful registration.
func (c *Client) RunRegisterWithContext(ctx context.Context, emailAddr, password, name, birthdate string) error {
	err := c.runFlow(ctx, emailAddr, password, name, birthdate)
	if err == nil && c.codexEnabled {
		if codexErr := c.extractCodexTokens(ctx, emailAddr); codexErr != nil {
			c.print("Codex extraction failed: " + codexErr.Error())
		}
	}
	return err
}

func applyTraceHeaders(req *http.Request) {
	for k, v := range util.MakeTraceHeaders() {
		req.Header.Set(k, v)
	}
}

func extractCallbackURL(data map[string]interface{}) string {
	for _, key := range []string{"continue_url", "url", "redirect_url"} {
		if nextURL, ok := data[key].(string); ok {
			return nextURL
		}
	}
	return ""
}

func isSuccessStatus(status int) bool {
	return status >= 200 && status < 400
}

// runFlow executes the registration state machine.
func (c *Client) runFlow(ctx context.Context, emailAddr, password, name, birthdate string) error {
	c.print("Starting registration flow...")

	if err := c.visitHomepage(); err != nil {
		return WrapFailure("visit_homepage", 0, err)
	}
	if err := c.randomDelayWithContext(ctx, 0.3, 0.8); err != nil {
		return WrapFailure("visit_homepage_delay", 0, err)
	}

	csrf, err := c.getCSRF()
	if err != nil {
		return WrapFailure("get_csrf", 0, err)
	}
	if err := c.randomDelayWithContext(ctx, 0.2, 0.5); err != nil {
		return WrapFailure("csrf_delay", 0, err)
	}

	authorizeURL, err := c.signin(emailAddr, csrf)
	if err != nil {
		return WrapFailure("signin", 0, err)
	}
	if err := c.randomDelayWithContext(ctx, 0.3, 0.8); err != nil {
		return WrapFailure("signin_delay", 0, err)
	}

	finalURL, err := c.authorize(authorizeURL)
	if err != nil {
		return WrapFailure("authorize", 0, err)
	}
	if err := c.randomDelayWithContext(ctx, 0.3, 0.8); err != nil {
		return WrapFailure("authorize_delay", 0, err)
	}

	u, _ := url.Parse(finalURL)
	finalPath := u.Path
	needOTP := false

	if strings.Contains(finalPath, "create-account/password") {
		if err := c.randomDelayWithContext(ctx, 0.5, 1.0); err != nil {
			return WrapFailure("pre_register_delay", 0, err)
		}
		status, data, runErr := c.register(emailAddr, password)
		if runErr != nil {
			return WrapFailure("register", status, runErr)
		}
		if !isSuccessStatus(status) {
			return NewFailure(classifyStatusFailure(status, data), "register", status, fmt.Errorf("register failed: %v", data))
		}
		if err := c.randomDelayWithContext(ctx, 0.3, 0.8); err != nil {
			return WrapFailure("post_register_delay", 0, err)
		}
		status, data, sendErr := c.sendOTP()
		if sendErr != nil {
			return WrapFailure("send_otp", status, sendErr)
		}
		if !isSuccessStatus(status) {
			return NewFailure(classifyStatusFailure(status, data), "send_otp", status, fmt.Errorf("send otp failed: %v", data))
		}
		needOTP = true
	} else if strings.Contains(finalPath, "email-verification") || strings.Contains(finalPath, "email-otp") {
		c.print("Jump to OTP verification stage")
		needOTP = true
	} else if strings.Contains(finalPath, "about-you") {
		c.print("Jump to fill information stage")
		if err := c.randomDelayWithContext(ctx, 0.5, 1.0); err != nil {
			return WrapFailure("pre_create_account_delay", 0, err)
		}
		_, data, runErr := c.createAccountWithPhoneRetry(ctx, name, birthdate, "create_account")
		if runErr != nil {
			return runErr
		}
		if err := c.randomDelayWithContext(ctx, 0.3, 0.5); err != nil {
			return WrapFailure("post_create_account_delay", 0, err)
		}

		cbURL := extractCallbackURL(data)
		status, data, callbackErr := c.callback(cbURL)
		if callbackErr != nil {
			return WrapFailure("callback", status, callbackErr)
		}
		if !isSuccessStatus(status) {
			return NewFailure(classifyStatusFailure(status, data), "callback", status, fmt.Errorf("callback failed: %v", data))
		}
		return nil
	} else if strings.Contains(finalPath, "callback") || strings.Contains(finalURL, "chatgpt.com") {
		c.print("Account registration completed")
		return nil
	} else {
		c.print(fmt.Sprintf("Unknown jump: %s", finalURL))
		status, data, runErr := c.register(emailAddr, password)
		if runErr != nil {
			return WrapFailure("register", status, runErr)
		}
		if !isSuccessStatus(status) {
			return NewFailure(classifyStatusFailure(status, data), "register", status, fmt.Errorf("register failed: %v", data))
		}
		status, data, sendErr := c.sendOTP()
		if sendErr != nil {
			return WrapFailure("send_otp", status, sendErr)
		}
		if !isSuccessStatus(status) {
			return NewFailure(classifyStatusFailure(status, data), "send_otp", status, fmt.Errorf("send otp failed: %v", data))
		}
		needOTP = true
	}

	if needOTP {
		otpCode, otpErr := c.otpProvider.GetOTP(ctx, emailAddr, defaultOTPTimeout)
		if otpErr != nil {
			return NewFailure(FailureOTPTimeout, "get_otp", 0, otpErr)
		}

		if err := c.randomDelayWithContext(ctx, 0.3, 0.8); err != nil {
			return WrapFailure("pre_validate_otp_delay", 0, err)
		}
		status, _, validateErr := c.validateOTP(otpCode)
		if validateErr != nil {
			return WrapFailure("validate_otp", status, validateErr)
		}

		if !isSuccessStatus(status) {
			c.print("Verification code failed, retrying...")
			status, data, sendErr := c.sendOTP()
			if sendErr != nil {
				return WrapFailure("retry_send_otp", status, sendErr)
			}
			if !isSuccessStatus(status) {
				return NewFailure(classifyStatusFailure(status, data), "retry_send_otp", status, fmt.Errorf("send otp failed: %v", data))
			}
			if err := c.randomDelayWithContext(ctx, 1.0, 2.0); err != nil {
				return WrapFailure("retry_otp_delay", 0, err)
			}
			otpCode, otpErr = c.otpProvider.GetOTP(ctx, emailAddr, 30*time.Second)
			if otpErr != nil {
				return NewFailure(FailureOTPTimeout, "retry_get_otp", status, otpErr)
			}
			if err := c.randomDelayWithContext(ctx, 0.3, 0.8); err != nil {
				return WrapFailure("retry_validate_otp_delay", 0, err)
			}
			status, data, validateErr = c.validateOTP(otpCode)
			if validateErr != nil {
				return WrapFailure("retry_validate_otp", status, validateErr)
			}
			if !isSuccessStatus(status) {
				return NewFailure(FailureOTPTimeout, "validate_otp", status, fmt.Errorf("verification code failed after retry: %v", data))
			}
		}
	}

	if err := c.randomDelayWithContext(ctx, 0.5, 1.5); err != nil {
		return WrapFailure("pre_final_create_account_delay", 0, err)
	}
	_, data, createErr := c.createAccountWithPhoneRetry(ctx, name, birthdate, "final_create_account")
	if createErr != nil {
		return createErr
	}

	if err := c.randomDelayWithContext(ctx, 0.2, 0.5); err != nil {
		return WrapFailure("pre_callback_delay", 0, err)
	}
	cbURL := extractCallbackURL(data)
	status, data, callbackErr := c.callback(cbURL)
	if callbackErr != nil {
		return WrapFailure("callback", status, callbackErr)
	}
	if !isSuccessStatus(status) {
		return NewFailure(classifyStatusFailure(status, data), "callback", status, fmt.Errorf("callback failed: %v", data))
	}

	return nil
}

func classifyStatusFailure(status int, data map[string]interface{}) FailureKind {
	if status == 429 {
		return FailureRateLimited
	}
	message := strings.ToLower(fmt.Sprintf("%v", data))
	if strings.Contains(message, "unsupported_email") {
		return FailureUnsupportedEmail
	}
	if isPhoneChallengeMessage(message) {
		return FailurePhoneChallenge
	}
	if strings.Contains(message, "challenge") || strings.Contains(message, "sentinel") {
		return FailureChallengeFailed
	}
	return FailureUpstreamChanged
}

func isPhoneChallengeMessage(message string) bool {
	hasPhoneSignal := strings.Contains(message, "phone") || strings.Contains(message, "sms") || strings.Contains(message, "otp")
	if !hasPhoneSignal {
		return false
	}
	return strings.Contains(message, "challenge") || strings.Contains(message, "verify") || strings.Contains(message, "verification")
}

func (c *Client) randomDelayWithContext(ctx context.Context, low, high float64) error {
	delay := low + rand.Float64()*(high-low)
	return waitWithContext(ctx, time.Duration(delay*float64(time.Second)))
}

// createAccountWithPhoneRetry calls createAccount and, on a phone challenge,
// handles the verification via phoneProvider before retrying.
func (c *Client) createAccountWithPhoneRetry(ctx context.Context, name, birthdate, step string) (int, map[string]interface{}, error) {
	status, data, err := c.createAccount(name, birthdate)
	if err != nil {
		return status, data, WrapFailure(step, status, err)
	}
	if !isSuccessStatus(status) {
		kind := classifyStatusFailure(status, data)
		if kind == FailurePhoneChallenge {
			if phErr := c.handlePhoneChallenge(ctx); phErr != nil {
				return status, data, phErr
			}
			// Retry createAccount after phone verification succeeds.
			status, data, err = c.createAccount(name, birthdate)
			if err != nil {
				return status, data, WrapFailure(step+"_retry", status, err)
			}
			if !isSuccessStatus(status) {
				return status, data, NewFailure(classifyStatusFailure(status, data), step+"_retry", status,
					fmt.Errorf("create account failed after phone verification: %v", data))
			}
		} else {
			return status, data, NewFailure(kind, step, status, fmt.Errorf("create account failed: %v", data))
		}
	}
	return status, data, nil
}

// handlePhoneChallenge rents a phone number via phoneProvider, submits it to OpenAI,
// waits for the SMS OTP, and validates it.
func (c *Client) handlePhoneChallenge(ctx context.Context) error {
	if c.phoneProvider == nil {
		return NewFailure(FailurePhoneChallenge, "phone_challenge", 0,
			fmt.Errorf("phone challenge detected but no phone provider configured (use --viotp-token)"))
	}
	result, err := c.phoneProvider.RentNumber(ctx, c.viOTPServiceID)
	if err != nil {
		return NewFailure(FailurePhoneRentFailed, "rent_phone", 0, err)
	}
	c.print(fmt.Sprintf("Phone rented: ...%s (requestID: %s)", maskPhone(result.PhoneNumber), result.RequestID))

	if err := c.submitPhoneNumber(ctx, result.PhoneNumber); err != nil {
		return NewFailure(FailurePhoneChallenge, "submit_phone", 0, err)
	}

	otp, err := c.phoneProvider.WaitForOTP(ctx, result.RequestID, phoneOTPTimeout)
	if err != nil {
		return NewFailure(FailurePhoneOTPTimeout, "wait_phone_otp", 0, err)
	}

	return c.validatePhoneOTP(ctx, otp)
}

// submitPhoneNumber sends the rented phone number to OpenAI's phone verification endpoint.
func (c *Client) submitPhoneNumber(ctx context.Context, phoneNumber string) error {
	payload := map[string]string{"phone": phoneNumber}
	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", phoneVerifyEndpoint, bytes.NewReader(jsonPayload))
	if err != nil {
		return fmt.Errorf("phone submit request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", authURL+"/create-account")
	applyTraceHeaders(req)

	resp, err := c.do(req)
	if err != nil {
		return fmt.Errorf("phone submit request failed: %w", err)
	}
	defer resp.Body.Close()

	c.log("Submit Phone", resp.StatusCode)
	if !isSuccessStatus(resp.StatusCode) {
		return fmt.Errorf("phone submit failed (status=%d)", resp.StatusCode)
	}
	return nil
}

// validatePhoneOTP submits the OTP code to OpenAI's phone validation endpoint.
func (c *Client) validatePhoneOTP(ctx context.Context, otp string) error {
	payload := map[string]string{"code": otp}
	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", phoneOTPEndpoint, bytes.NewReader(jsonPayload))
	if err != nil {
		return fmt.Errorf("phone OTP validation request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", authURL+"/phone-verification")
	applyTraceHeaders(req)

	resp, err := c.do(req)
	if err != nil {
		return fmt.Errorf("phone OTP validation request failed: %w", err)
	}
	defer resp.Body.Close()

	c.log("Validate Phone OTP", resp.StatusCode)
	if !isSuccessStatus(resp.StatusCode) {
		return NewFailure(FailurePhoneOTPTimeout, "validate_phone_otp", resp.StatusCode,
			fmt.Errorf("phone OTP validation failed (status=%d)", resp.StatusCode))
	}
	return nil
}

// extractCodexTokens runs the Codex PKCE OAuth flow immediately after a successful
// registration using the existing authenticated session, then appends the tokens to
// the output file. Returns an error on failure; does not affect the registration result.
func (c *Client) extractCodexTokens(ctx context.Context, emailAddr string) error {
	pkce, err := codex.GeneratePKCE()
	if err != nil {
		return fmt.Errorf("codex: PKCE generation failed: %w", err)
	}

	state := util.GenerateUUID()

	// StartCallbackInterceptor binds the listener synchronously, eliminating the
	// timing race from a fixed port + sleep approach, and gives each worker its own
	// ephemeral port to avoid "bind: address already in use" under concurrent workers.
	redirectURI, waitFn, err := codex.StartCallbackInterceptor(ctx, state, codexCallbackTimeout)
	if err != nil {
		return fmt.Errorf("codex: %w", err)
	}

	cfg := codex.DefaultSSOConfig()
	cfg.RedirectURI = redirectURI
	authorizeURL := codex.BuildAuthorizeURL(cfg, pkce, state)

	codeCh := make(chan string, 1)
	cbErrCh := make(chan error, 1)
	go func() {
		code, waitErr := waitFn()
		if waitErr != nil {
			cbErrCh <- waitErr
		} else {
			codeCh <- code
		}
	}()

	req, reqErr := http.NewRequestWithContext(ctx, "GET", authorizeURL, nil)
	if reqErr != nil {
		return fmt.Errorf("codex: create authorize request: %w", reqErr)
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	resp, err := c.do(req)
	if err != nil {
		return fmt.Errorf("codex: authorize request failed: %w", err)
	}
	if resp != nil {
		resp.Body.Close()
	}

	select {
	case code := <-codeCh:
		tokens, exchErr := codex.ExchangeCode(ctx, cfg, code, pkce.Verifier)
		if exchErr != nil {
			return fmt.Errorf("codex: token exchange failed: %w", exchErr)
		}
		c.fileMu.Lock()
		writeErr := writeCodexTokens(c.codexOutput, emailAddr, tokens)
		c.fileMu.Unlock()
		if writeErr != nil {
			return writeErr
		}
		c.print("Codex: tokens saved to " + c.codexOutput)
		return nil
	case cbErr := <-cbErrCh:
		return fmt.Errorf("codex: %w", cbErr)
	case <-ctx.Done():
		return ctx.Err()
	}
}

// maskPhone returns only the last 4 digits of a phone number to avoid PII in logs.
func maskPhone(phone string) string {
	if len(phone) <= 4 {
		return "****"
	}
	return phone[len(phone)-4:]
}

type codexTokenEntry struct {
	Email        string `json:"email"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresAt    string `json:"expires_at"`
	CreatedAt    string `json:"created_at"`
}

// writeCodexTokens appends a new token entry to the JSON array output file.
// Must be called under fileMu lock. Uses atomic temp-file + rename for safety.
func writeCodexTokens(outputFile, emailAddr string, tokens *codex.TokenResult) error {
	var entries []codexTokenEntry
	if existing, readErr := os.ReadFile(outputFile); readErr == nil {
		_ = json.Unmarshal(existing, &entries)
	}

	now := time.Now().UTC()
	entries = append(entries, codexTokenEntry{
		Email:        emailAddr,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    now.Add(time.Duration(tokens.ExpiresIn) * time.Second).Format(time.RFC3339),
		CreatedAt:    now.Format(time.RFC3339),
	})

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("codex: marshal tokens: %w", err)
	}

	tmp := outputFile + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("codex: write token file: %w", err)
	}
	if err := os.Rename(tmp, outputFile); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("codex: rename token file: %w", err)
	}
	return nil
}
