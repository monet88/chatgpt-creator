package email

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v7"

	"github.com/PuerkitoBio/goquery"
	fhttp "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"

	"github.com/monet88/chatgpt-creator/internal/chrome"
	"github.com/monet88/chatgpt-creator/internal/util"
)

var (
	blacklistedDomains sync.Map
	blacklistMutex     sync.Mutex
)

func init() {
	data, err := os.ReadFile("blacklist.json")
	if err != nil {
		return // File might not exist yet
	}

	var domains []string
	if err := json.Unmarshal(data, &domains); err != nil {
		return
	}

	for _, domain := range domains {
		blacklistedDomains.Store(domain, true)
	}
}

func saveBlacklist() {
	blacklistMutex.Lock()
	defer blacklistMutex.Unlock()

	var domains []string
	blacklistedDomains.Range(func(key, value any) bool {
		if domain, ok := key.(string); ok {
			domains = append(domains, domain)
		}
		return true
	})

	data, err := json.MarshalIndent(domains, "", "  ")
	if err != nil {
		return
	}

	_ = os.WriteFile("blacklist.json", data, 0644)
}

// AddBlacklistDomain adds a domain to the global blacklist.
func AddBlacklistDomain(domain string) {
	blacklistedDomains.Store(domain, true)
	saveBlacklist()
}

// CreateTempEmail fetches a new temp email using a random profile and gofakeit names.
func CreateTempEmail(defaultDomain string) (string, error) {
	// If defaultDomain is set, skip fetching from generator.email
	if defaultDomain != "" {
		firstName := gofakeit.FirstName()
		lastName := gofakeit.LastName()
		email := strings.ToLower(firstName+lastName+util.RandStr(5)) + "@" + defaultDomain
		return email, nil
	}

	tlsProfile := randomTLSProfile()
	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(tlsProfile),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return "", fmt.Errorf("failed to create tls client: %w", err)
	}

	req, err := fhttp.NewRequest(http.MethodGet, "https://generator.email/", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch generator.email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("generator.email returned status: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	domains := []string{"smartmail.de", "enayu.com", "crazymailing.com"}
	doc.Find(".e7m.tt-suggestions div > p").Each(func(i int, s *goquery.Selection) {
		domain := strings.TrimSpace(s.Text())
		if domain != "" {
			if _, blacklisted := blacklistedDomains.Load(domain); !blacklisted {
				domains = append(domains, domain)
			}
		}
	})

	if len(domains) == 0 {
		return "", fmt.Errorf("all available domains are blacklisted")
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomDomain := domains[r.Intn(len(domains))]

	firstName := gofakeit.FirstName()
	lastName := gofakeit.LastName()
	var localPart string
	switch r.Intn(5) {
	case 0:
		localPart = strings.ToLower(firstName + lastName)
	case 1:
		localPart = strings.ToLower(firstName + "." + lastName)
	case 2:
		localPart = strings.ToLower(lastName + firstName)
	case 3:
		localPart = strings.ToLower(firstName)
	case 4:
		localPart = strings.ToLower(lastName)
	}
	localPart += util.RandStr(3 + r.Intn(4))
	email := localPart + "@" + randomDomain

	return email, nil
}

var otpRegex = regexp.MustCompile(`\d{6}`)

func parseVerificationCodeFromHTML(reader io.Reader) (string, error) {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return "", err
	}

	otp := ""
	doc.Find("#email-table > div.e7m.list-group-item.list-group-item-info > div.e7m.subj_div_45g45gg").EachWithBreak(func(i int, s *goquery.Selection) bool {
		text := s.Text()
		matches := otpRegex.FindStringSubmatch(text)
		if len(matches) == 0 {
			return true
		}
		code := matches[0]
		if code == "177010" {
			return true
		}
		otp = code
		return false
	})

	return otp, nil
}

// randomTLSProfile returns a random Chrome TLS profile for fingerprint diversity.
func randomTLSProfile() profiles.ClientProfile {
	profile, _, _ := chrome.RandomChromeVersion()
	return chrome.MapToTLSProfile(profile.Impersonate)
}

// GetVerificationCode polls generator.email for the OTP using a custom cookie.
func GetVerificationCode(email string, maxRetries int, delay time.Duration) (string, error) {
	return GetVerificationCodeWithContext(context.Background(), email, maxRetries, delay)
}

// GetVerificationCodeWithContext polls generator.email for the OTP using a custom cookie and context-aware waits.
func GetVerificationCodeWithContext(ctx context.Context, email string, maxRetries int, delay time.Duration) (string, error) {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid email format: %s", email)
	}
	username := parts[0]
	domain := parts[1]

	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		options := []tls_client.HttpClientOption{
			tls_client.WithClientProfile(randomTLSProfile()),
		}

		client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
		if err != nil {
			return "", fmt.Errorf("failed to create tls client: %w", err)
		}

		url := fmt.Sprintf("https://generator.email/%s/%s", domain, username)
		req, err := fhttp.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Cookie", fmt.Sprintf("surl=%s/%s", domain, username))

		resp, err := client.Do(req)
		if err != nil {
			if util.WaitWithContext(ctx, delay) != nil {
				return "", ctx.Err()
			}
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			if util.WaitWithContext(ctx, delay) != nil {
				return "", ctx.Err()
			}
			continue
		}

		otp, err := parseVerificationCodeFromHTML(resp.Body)
		resp.Body.Close()
		if err != nil {
			if util.WaitWithContext(ctx, delay) != nil {
				return "", ctx.Err()
			}
			continue
		}

		if otp != "" {
			return otp, nil
		}

		if util.WaitWithContext(ctx, delay) != nil {
			return "", ctx.Err()
		}
	}

	return "", fmt.Errorf("failed to get verification code after %d retries", maxRetries)
}
