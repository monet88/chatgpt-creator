package chrome

import (
	"fmt"
	"math/rand"

	"github.com/bogdanfinn/tls-client/profiles"
)

// Profile represents a Chrome browser fingerprint variant.
type Profile struct {
	Major       int
	Impersonate string
	Build       int
	PatchMin    int
	PatchMax    int
	SecChUA     string
	Platform    string // "Windows", "macOS", or "Linux"
}

// chromeProfiles defines all supported Chrome TLS impersonation profiles.
// Each profile maps to a tls-client ClientProfile for JA3/JA4 fingerprint diversity.
var chromeProfiles = []Profile{
	{103, "chrome103", 5060, 0, 134, "\"Chromium\";v=\"103\", \"Google Chrome\";v=\"103\", \"Not_A Brand\";v=\"99\"", ""},
	{104, "chrome104", 5112, 0, 146, "\"Chromium\";v=\"104\", \"Google Chrome\";v=\"104\", \"Not_A Brand\";v=\"99\"", ""},
	{105, "chrome105", 5195, 0, 134, "\"Chromium\";v=\"105\", \"Google Chrome\";v=\"105\", \"Not_A Brand\";v=\"99\"", ""},
	{106, "chrome106", 5249, 0, 159, "\"Chromium\";v=\"106\", \"Google Chrome\";v=\"106\", \"Not_A Brand\";v=\"99\"", ""},
	{107, "chrome107", 5304, 0, 141, "\"Chromium\";v=\"107\", \"Google Chrome\";v=\"107\", \"Not_A Brand\";v=\"99\"", ""},
	{108, "chrome108", 5359, 0, 141, "\"Chromium\";v=\"108\", \"Google Chrome\";v=\"108\", \"Not_A Brand\";v=\"99\"", ""},
	{109, "chrome109", 5414, 0, 120, "\"Chromium\";v=\"109\", \"Google Chrome\";v=\"109\", \"Not_A Brand\";v=\"99\"", ""},
	{110, "chrome110", 5481, 0, 177, "\"Chromium\";v=\"110\", \"Google Chrome\";v=\"110\", \"Not_A Brand\";v=\"99\"", ""},
	{111, "chrome111", 5563, 0, 146, "\"Chromium\";v=\"111\", \"Google Chrome\";v=\"111\", \"Not_A Brand\";v=\"99\"", ""},
	{112, "chrome112", 5615, 0, 199, "\"Chromium\";v=\"112\", \"Google Chrome\";v=\"112\", \"Not_A Brand\";v=\"99\"", ""},
	{116, "chrome116", 5845, 0, 177, "\"Chromium\";v=\"116\", \"Google Chrome\";v=\"116\", \"Not_A Brand\";v=\"8\"", ""},
	{117, "chrome117", 5938, 0, 176, "\"Chromium\";v=\"117\", \"Google Chrome\";v=\"117\", \"Not_A Brand\";v=\"8\"", ""},
	{120, "chrome120", 6099, 0, 147, "\"Chromium\";v=\"120\", \"Google Chrome\";v=\"120\", \"Not_A Brand\";v=\"8\"", ""},
	{124, "chrome124", 6367, 0, 205, "\"Chromium\";v=\"124\", \"Google Chrome\";v=\"124\", \"Not_A Brand\";v=\"24\"", ""},
	{130, "chrome130", 6723, 0, 126, "\"Chromium\";v=\"130\", \"Google Chrome\";v=\"130\", \"Not_A Brand\";v=\"24\"", ""},
	{131, "chrome131", 6778, 0, 204, "\"Chromium\";v=\"131\", \"Google Chrome\";v=\"131\", \"Not_A Brand\";v=\"24\"", ""},
	{133, "chrome133", 6943, 0, 141, "\"Chromium\";v=\"133\", \"Google Chrome\";v=\"133\", \"Not_A Brand\";v=\"24\"", ""},
	{144, "chrome144", 7569, 0, 100, "\"Chromium\";v=\"144\", \"Google Chrome\";v=\"144\", \"Not_A Brand\";v=\"24\"", ""},
	{146, "chrome146", 7636, 0, 100, "\"Chromium\";v=\"146\", \"Google Chrome\";v=\"146\", \"Not_A Brand\";v=\"24\"", ""},
}

// platformDefinitions maps platform codes to OS identifiers for User-Agent and sec-ch-ua-platform.
var platformDefinitions = []struct {
	Key      string
	OSIdent  string // Used in sec-ch-ua-platform
	UAPrefix string // Full OS prefix for User-Agent string
	Weight   int
}{
	{"Windows", `"Windows"`, "Windows NT 10.0; Win64; x64", 70},
	{"macOS", `"macOS"`, "Macintosh; Intel Mac OS X 10_15_7", 25},
	{"Linux", `"Linux"`, "X11; Linux x86_64", 5},
}

// RandomChromeVersion selects a random Chrome profile and OS platform,
// returning the profile, the full version string, and the User-Agent header.
func RandomChromeVersion() (Profile, string, string) {
	profile := chromeProfiles[rand.Intn(len(chromeProfiles))]

	// Pick random platform weighted by real-world distribution
	platform := pickWeightedPlatform()
	profile.Platform = platform.Key

	patch := rand.Intn(profile.PatchMax-profile.PatchMin+1) + profile.PatchMin
	fullVersion := fmt.Sprintf("%d.0.%d.%d", profile.Major, profile.Build, patch)
	userAgent := fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36",
		platform.UAPrefix, fullVersion)
	return profile, fullVersion, userAgent
}

// pickWeightedPlatform selects a random OS platform based on configured weights.
func pickWeightedPlatform() struct {
	Key      string
	OSIdent  string
	UAPrefix string
	Weight   int
} {
	totalWeight := 0
	for _, p := range platformDefinitions {
		totalWeight += p.Weight
	}
	r := rand.Intn(totalWeight)
	for _, p := range platformDefinitions {
		r -= p.Weight
		if r < 0 {
			return p
		}
	}
	return platformDefinitions[0]
}

// MapToTLSProfile maps an impersonate string to the corresponding tls-client ClientProfile.
func MapToTLSProfile(impersonate string) profiles.ClientProfile {
	switch impersonate {
	case "chrome103":
		return profiles.Chrome_103
	case "chrome104":
		return profiles.Chrome_104
	case "chrome105":
		return profiles.Chrome_105
	case "chrome106":
		return profiles.Chrome_106
	case "chrome107":
		return profiles.Chrome_107
	case "chrome108":
		return profiles.Chrome_108
	case "chrome109":
		return profiles.Chrome_109
	case "chrome110":
		return profiles.Chrome_110
	case "chrome111":
		return profiles.Chrome_111
	case "chrome112":
		return profiles.Chrome_112
	case "chrome116":
		return profiles.Chrome_116_PSK
	case "chrome117":
		return profiles.Chrome_117
	case "chrome120":
		return profiles.Chrome_120
	case "chrome124":
		return profiles.Chrome_124
	case "chrome130":
		return profiles.Chrome_130_PSK
	case "chrome131":
		return profiles.Chrome_131
	case "chrome133", "chrome133a":
		return profiles.Chrome_133
	case "chrome136", "chrome142":
		// Fallback to Chrome_133 as these are not available in tls-client v1.14.0
		return profiles.Chrome_133
	case "chrome144":
		return profiles.Chrome_144
	case "chrome146":
		return profiles.Chrome_146
	default:
		return profiles.Chrome_133
	}
}

// PlatformOSIdent returns the sec-ch-ua-platform value for a platform key.
func PlatformOSIdent(platform string) string {
	for _, p := range platformDefinitions {
		if p.Key == platform {
			return p.OSIdent
		}
	}
	return `"Windows"`
}
