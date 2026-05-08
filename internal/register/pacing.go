package register

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// PacingProfile controls the inter-registration delay between consecutive attempts.
type PacingProfile string

const (
	// PacingNone disables inter-registration delays (testing/debugging only).
	PacingNone PacingProfile = "none"
	// PacingFast uses 5-15s delays (suitable for rotating residential proxies).
	PacingFast PacingProfile = "fast"
	// PacingHuman uses 120-300s delays (production default, mimics human pacing).
	PacingHuman PacingProfile = "human"
	// PacingSlow uses 300-600s delays (maximum safety for single IP).
	PacingSlow PacingProfile = "slow"
)

type pacingRange struct {
	minSeconds float64
	maxSeconds float64
}

var pacingRanges = map[PacingProfile]pacingRange{
	PacingNone:  {0, 0},
	PacingFast:  {5, 15},
	PacingHuman: {120, 300},
	PacingSlow:  {300, 600},
}

// ParsePacingProfile validates and returns a PacingProfile from a string input.
func ParsePacingProfile(s string) (PacingProfile, error) {
	profile := PacingProfile(strings.ToLower(strings.TrimSpace(s)))
	if _, ok := pacingRanges[profile]; !ok {
		return "", fmt.Errorf("invalid pacing profile %q: must be one of none, fast, human, slow", s)
	}
	return profile, nil
}

// pacingDelay returns a random duration within the pacing profile's range.
// Returns zero for PacingNone.
func pacingDelay(profile PacingProfile) time.Duration {
	r, ok := pacingRanges[profile]
	if !ok || r.maxSeconds == 0 {
		return 0
	}
	seconds := r.minSeconds + rand.Float64()*(r.maxSeconds-r.minSeconds)
	return time.Duration(seconds * float64(time.Second))
}
