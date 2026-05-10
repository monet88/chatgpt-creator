package register

import (
	"fmt"
	"math/rand"
	"net/url"
	"sync"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/tls-client"
	"github.com/google/uuid"

	"github.com/monet88/chatgpt-creator/internal/chrome"
	"github.com/monet88/chatgpt-creator/internal/email"
	"github.com/monet88/chatgpt-creator/internal/phone"
)

const (
	baseURL = "https://chatgpt.com"
	authURL = "https://auth.openai.com"
)

type Client struct {
	session        tls_client.HttpClient
	proxy          string
	tag            string
	workerID       int
	deviceID       string
	impersonate    string
	major          int
	fullVersion    string
	ua             string
	secChUA        string
	platform       string
	acceptLanguage string
	acceptEncoding string
	printMu        *sync.Mutex
	fileMu         *sync.Mutex
	otpProvider    email.OTPProvider
	phoneProvider  phone.PhoneProvider
	viOTPServiceID int
	codexEnabled   bool
	codexOutput    string
	panelOutputDir string
	mfaEnabled      bool
	camofoxURL      string
	totpSecret      string
	accountPassword string
}

func NewClient(proxy, tag string, workerID int, printMu, fileMu *sync.Mutex) (*Client, error) {
	profile, fullVersion, ua := chrome.RandomChromeVersion()
	impersonate := profile.Impersonate
	mappedProfile := chrome.MapToTLSProfile(impersonate)

	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(mappedProfile),
		tls_client.WithCookieJar(tls_client.NewCookieJar()),
	}

	if proxy != "" {
		options = append(options, tls_client.WithProxyUrl(proxy))
	}

	session, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	deviceID := uuid.New().String()

	c := &Client{
		session:     session,
		proxy:       proxy,
		tag:         tag,
		workerID:    workerID,
		deviceID:    deviceID,
		impersonate: impersonate,
		fullVersion: fullVersion,
		ua:          ua,
		printMu:     printMu,
		fileMu:      fileMu,
	}

	// major version for sec-ch-ua
	c.major = profile.Major
	c.secChUA = profile.SecChUA

	// Randomize browser attributes for fingerprint diversity
	c.platform = profile.Platform
	c.acceptLanguage = randomAcceptLanguage()
	c.acceptEncoding = randomAcceptEncoding()

	// Add initial cookie
	u, _ := url.Parse(baseURL)
	cookies := []*http.Cookie{
		{
			Name:   "oai-did",
			Value:  deviceID,
			Domain: "chatgpt.com",
			Path:   "/",
		},
	}
	session.GetCookieJar().SetCookies(u, cookies)

	return c, nil
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	// Build a randomized list of header setter functions to avoid fingerprinting
	type headerSetter struct {
		key   string
		value string
	}
	var setters []headerSetter

	if req.Header.Get("User-Agent") == "" {
		setters = append(setters, headerSetter{"User-Agent", c.ua})
	}
	if req.Header.Get("Accept") == "" {
		setters = append(setters, headerSetter{"Accept", "*/*"})
	}
	if req.Header.Get("Accept-Language") == "" {
		setters = append(setters, headerSetter{"Accept-Language", c.acceptLanguage})
	}
	if req.Header.Get("Accept-Encoding") == "" {
		setters = append(setters, headerSetter{"Accept-Encoding", c.acceptEncoding})
	}
	if req.Header.Get("sec-ch-ua") == "" {
		setters = append(setters, headerSetter{"sec-ch-ua", c.secChUA})
	}
	if req.Header.Get("sec-ch-ua-mobile") == "" {
		setters = append(setters, headerSetter{"sec-ch-ua-mobile", "?0"})
	}
	if req.Header.Get("sec-ch-ua-platform") == "" {
		setters = append(setters, headerSetter{"sec-ch-ua-platform", chrome.PlatformOSIdent(c.platform)})
	}

	// Shuffle header order for fingerprint diversity
	rand.Shuffle(len(setters), func(i, j int) {
		setters[i], setters[j] = setters[j], setters[i]
	})

	for _, s := range setters {
		req.Header.Set(s.key, s.value)
	}

	return c.session.Do(req)
}

func (c *Client) log(step string, status int) {
	c.printMu.Lock()
	defer c.printMu.Unlock()

	ts := time.Now().Format("15:04:05")
	diagnosticPrintf("[%s] [W%d] [%s] %s | %d\n", ts, c.workerID, c.tag, safeLogMessage(step), status)
}

func (c *Client) print(msg string) {
	c.printMu.Lock()
	defer c.printMu.Unlock()

	ts := time.Now().Format("15:04:05")
	diagnosticPrintf("[%s] [W%d] [%s] %s\n", ts, c.workerID, c.tag, safeLogMessage(msg))
}

// acceptLanguageOptions provides diverse Accept-Language header values.
var acceptLanguageOptions = []string{
	"en-US,en;q=0.9",
	"en-US,en;q=0.9,vi;q=0.8",
	"en-GB,en;q=0.9",
	"en-US,en;q=0.9,es;q=0.8",
	"en-US,en;q=0.9,fr;q=0.8",
	"en;q=0.9",
	"en-US,en;q=0.8",
	"en-CA,en;q=0.9",
}

// acceptEncodingOptions provides diverse Accept-Encoding header values.
var acceptEncodingOptions = []string{
	"gzip, deflate, br",
	"gzip, deflate, br, zstd",
	"gzip, deflate",
}

func randomAcceptLanguage() string {
	return acceptLanguageOptions[rand.Intn(len(acceptLanguageOptions))]
}

func randomAcceptEncoding() string {
	return acceptEncodingOptions[rand.Intn(len(acceptEncodingOptions))]
}

func (c *Client) TOTPSecret() string {
	return c.totpSecret
}
