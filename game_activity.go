package micup

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// Game activity states matching Fluxer engine desktop client.
const (
	GameActivityStatePlaying     = "playing"
	GameActivityStateIdle        = "idle"
	GameActivityStateDisabled    = "disabled"
	GameActivityStateUnsupported = "unsupported"
	GameActivityStateUnavailable = "unavailable"
)

var (
	gameIDRegex    = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,47}$`)
	ruleIDRegex    = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,128}$`)
	blacklistedExe = regexp.MustCompile(`(?i)(?:^javaw?\.exe$|launcher|patcher|anti.?cheat|^steam\.exe$|^riotclientservices\.exe$)`)
)

// ValidateGameID validates the slug format of a game identifier.
func ValidateGameID(id string) bool {
	return gameIDRegex.MatchString(id)
}

// ValidateManualName checks that the custom title is 1..64 chars, <= 256 bytes, and has no control characters.
func ValidateManualName(name string) bool {
	trimmed := strings.TrimSpace(name)
	if len(trimmed) == 0 {
		return false
	}
	runes := []rune(trimmed)
	if len(runes) > 64 || len([]byte(trimmed)) > 256 {
		return false
	}
	for _, r := range runes {
		if unicode.Is(unicode.Cc, r) || unicode.Is(unicode.Cf, r) || unicode.Is(unicode.Cs, r) {
			return false
		}
	}
	return true
}

// ValidateExecutable checks if an executable name is acceptable for game detection rules.
func ValidateExecutable(exe string) bool {
	lower := strings.ToLower(exe)
	if len(lower) > 260 || !strings.HasSuffix(lower, ".exe") {
		return false
	}
	if strings.ContainsAny(lower, `\/:*?"<>|`) {
		return false
	}
	return !blacklistedExe.MatchString(lower)
}

// GameActivityDetails describes the active game metadata.
type GameActivityDetails struct {
	StartedAt  int64  `json:"started_at"`
	GameID     string `json:"game_id,omitempty"`
	ManualName string `json:"manual_name,omitempty"`
}

// GameActivityPayload represents the OpPresenceUpdate activity payload.
type GameActivityPayload struct {
	Generation int                  `json:"generation"`
	Fresh      bool                 `json:"fresh"`
	State      string               `json:"state"`
	Activity   *GameActivityDetails `json:"activity"`
}

// GameActivityRule maps a local process to a known game.
type GameActivityRule struct {
	RuleID     string `json:"rule_id"`
	Executable string `json:"exe"`
	GameID     string `json:"game_id,omitempty"`
	ManualName string `json:"manual_name,omitempty"`
}

// NewGameActivityPayload creates and validates a game activity payload.
func NewGameActivityPayload(gameID, manualName string, startedAt time.Time) (*GameActivityPayload, error) {
	if gameID == "" && manualName == "" {
		return nil, errors.New("either gameID or manualName must be provided")
	}

	activity := &GameActivityDetails{
		StartedAt: startedAt.UnixMilli(),
	}

	if gameID != "" {
		if !ValidateGameID(gameID) {
			return nil, fmt.Errorf("invalid game_id format %q", gameID)
		}
		activity.GameID = gameID
	} else {
		if !ValidateManualName(manualName) {
			return nil, fmt.Errorf("invalid manual_name %q (must be 1-64 chars, no control chars)", manualName)
		}
		activity.ManualName = manualName
	}

	return &GameActivityPayload{
		Generation: 1,
		Fresh:      true,
		State:      GameActivityStatePlaying,
		Activity:   activity,
	}, nil
}

// UpdatePlayingGame broadcasts an active gaming status to the gateway.
func (s *Session) UpdatePlayingGame(gameID, manualName string, startedAt time.Time) error {
	payload, err := NewGameActivityPayload(gameID, manualName, startedAt)
	if err != nil {
		return err
	}

	if s.gateway == nil {
		return errors.New("gateway connection is not active")
	}

	activityName := manualName
	if activityName == "" {
		activityName = gameID
	}

	_ = payload // validates schema
	return s.UpdateGameStatus(StatusOnline, activityName)
}

// ClearGameActivity clears active game activity on the gateway.
func (s *Session) ClearGameActivity() error {
	return s.UpdateStatus(StatusOnline, nil)
}
