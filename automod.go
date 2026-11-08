package micup

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// AutoMod Event Types.
const (
	AutoModEventMessageSend = 1
)

// AutoMod Trigger Types.
const (
	AutoModTriggerKeyword       = 1
	AutoModTriggerHarmfulLink   = 2
	AutoModTriggerSpam          = 3
	AutoModTriggerKeywordPreset = 4
	AutoModTriggerMentionSpam   = 5
)

// AutoMod Action Types.
const (
	AutoModActionBlockMessage     = 1
	AutoModActionSendAlertMessage = 2
	AutoModActionTimeout          = 3
)

// AutoMod Keyword Presets.
const (
	KeywordPresetProfanity     = 1
	KeywordPresetSexualContent = 2
	KeywordPresetSlurs         = 3
)

// AutoModTriggerMetadata defines metadata for matching content.
type AutoModTriggerMetadata struct {
	KeywordFilter               []string `json:"keyword_filter,omitempty"`
	RegexPatterns               []string `json:"regex_patterns,omitempty"`
	Presets                     []int    `json:"presets,omitempty"`
	AllowList                   []string `json:"allow_list,omitempty"`
	MentionTotalLimit           int      `json:"mention_total_limit,omitempty"`
	MentionRaidProtectionEnabled bool     `json:"mention_raid_protection_enabled,omitempty"`
}

// AutoModActionMetadata defines behavior when an action is triggered.
type AutoModActionMetadata struct {
	ChannelID       string `json:"channel_id,omitempty"`
	DurationSeconds int    `json:"duration_seconds,omitempty"`
	CustomMessage   string `json:"custom_message,omitempty"`
}

// AutoModAction defines an action taken when a rule triggers.
type AutoModAction struct {
	Type     int                    `json:"type"`
	Metadata *AutoModActionMetadata `json:"metadata,omitempty"`
}

// AutoModerationRule represents a guild automoderation rule.
type AutoModerationRule struct {
	ID              string                  `json:"id"`
	GuildID         string                  `json:"guild_id"`
	Name            string                  `json:"name"`
	CreatorID       string                  `json:"creator_id"`
	EventType       int                     `json:"event_type"`
	TriggerType     int                     `json:"trigger_type"`
	TriggerMetadata *AutoModTriggerMetadata `json:"trigger_metadata"`
	Actions         []*AutoModAction        `json:"actions"`
	Enabled         bool                    `json:"enabled"`
	ExemptRoles     []string                `json:"exempt_roles"`
	ExemptChannels  []string                `json:"exempt_channels"`
}

// CreateAutoModRuleParams defines parameters for creating an AutoMod rule.
type CreateAutoModRuleParams struct {
	Name            string                  `json:"name"`
	EventType       int                     `json:"event_type"`
	TriggerType     int                     `json:"trigger_type"`
	TriggerMetadata *AutoModTriggerMetadata `json:"trigger_metadata"`
	Actions         []*AutoModAction        `json:"actions"`
	Enabled         *bool                   `json:"enabled,omitempty"`
	ExemptRoles     []string                `json:"exempt_roles,omitempty"`
	ExemptChannels  []string                `json:"exempt_channels,omitempty"`
}

// ModifyAutoModRuleParams defines parameters for modifying an existing AutoMod rule.
type ModifyAutoModRuleParams struct {
	Name            *string                 `json:"name,omitempty"`
	EventType       *int                    `json:"event_type,omitempty"`
	TriggerMetadata *AutoModTriggerMetadata `json:"trigger_metadata,omitempty"`
	Actions         *[]*AutoModAction       `json:"actions,omitempty"`
	Enabled         *bool                   `json:"enabled,omitempty"`
	ExemptRoles     *[]string               `json:"exempt_roles,omitempty"`
	ExemptChannels  *[]string               `json:"exempt_channels,omitempty"`
}

// RESTClient AutoMod methods

// AutoModerationRules lists all AutoMod rules for a guild.
func (c *RESTClient) AutoModerationRules(ctx context.Context, guildID string) ([]*AutoModerationRule, error) {
	var rules []*AutoModerationRule
	err := c.RequestJSON(ctx, "GET", fmt.Sprintf("/guilds/%s/auto-moderation/rules", guildID), nil, &rules)
	if err != nil {
		return nil, err
	}
	return rules, nil
}

// AutoModerationRule fetches a single rule by ID.
func (c *RESTClient) AutoModerationRule(ctx context.Context, guildID, ruleID string) (*AutoModerationRule, error) {
	var rule AutoModerationRule
	err := c.RequestJSON(ctx, "GET", fmt.Sprintf("/guilds/%s/auto-moderation/rules/%s", guildID, ruleID), nil, &rule)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// AutoModerationRuleCreate creates a new AutoMod rule.
func (c *RESTClient) AutoModerationRuleCreate(ctx context.Context, guildID string, params *CreateAutoModRuleParams) (*AutoModerationRule, error) {
	var rule AutoModerationRule
	err := c.RequestJSON(ctx, "POST", fmt.Sprintf("/guilds/%s/auto-moderation/rules", guildID), params, &rule)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// AutoModerationRuleModify updates an existing rule.
func (c *RESTClient) AutoModerationRuleModify(ctx context.Context, guildID, ruleID string, params *ModifyAutoModRuleParams) (*AutoModerationRule, error) {
	var rule AutoModerationRule
	err := c.RequestJSON(ctx, "PATCH", fmt.Sprintf("/guilds/%s/auto-moderation/rules/%s", guildID, ruleID), params, &rule)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// AutoModerationRuleDelete removes a rule.
func (c *RESTClient) AutoModerationRuleDelete(ctx context.Context, guildID, ruleID string) error {
	return c.RequestJSON(ctx, "DELETE", fmt.Sprintf("/guilds/%s/auto-moderation/rules/%s", guildID, ruleID), nil, nil)
}

// Session convenience methods

// AutoModerationRules lists guild rules.
func (s *Session) AutoModerationRules(guildID string) ([]*AutoModerationRule, error) {
	return s.rest.AutoModerationRules(context.Background(), guildID)
}

// AutoModerationRule fetches one rule.
func (s *Session) AutoModerationRule(guildID, ruleID string) (*AutoModerationRule, error) {
	return s.rest.AutoModerationRule(context.Background(), guildID, ruleID)
}

// AutoModerationRuleCreate creates a rule.
func (s *Session) AutoModerationRuleCreate(guildID string, params *CreateAutoModRuleParams) (*AutoModerationRule, error) {
	return s.rest.AutoModerationRuleCreate(context.Background(), guildID, params)
}

// AutoModerationRuleModify updates a rule.
func (s *Session) AutoModerationRuleModify(guildID, ruleID string, params *ModifyAutoModRuleParams) (*AutoModerationRule, error) {
	return s.rest.AutoModerationRuleModify(context.Background(), guildID, ruleID, params)
}

// AutoModerationRuleDelete deletes a rule.
func (s *Session) AutoModerationRuleDelete(guildID, ruleID string) error {
	return s.rest.AutoModerationRuleDelete(context.Background(), guildID, ruleID)
}

// -------------------------------------------------------------
// In-Memory Bot-Side AutoMod Engine
// -------------------------------------------------------------

// LocalAutoModRule represents an in-memory moderation filter evaluated in real time.
type LocalAutoModRule struct {
	Name        string
	Keywords    []string
	Regexes     []*regexp.Regexp
	MaxMentions int
	Action      func(s *Session, msg *Message, reason string)
}

// AutoModEngine evaluates messages locally against registered rules.
type AutoModEngine struct {
	rules []*LocalAutoModRule
	mu    sync.RWMutex
}

// NewAutoModEngine initializes a bot-side AutoMod engine.
func NewAutoModEngine() *AutoModEngine {
	return &AutoModEngine{
		rules: make([]*LocalAutoModRule, 0),
	}
}

// AddKeywordRule adds a rule that triggers when any of the keywords appear in the message.
func (e *AutoModEngine) AddKeywordRule(name string, keywords []string, action func(s *Session, msg *Message, reason string)) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.rules = append(e.rules, &LocalAutoModRule{
		Name:     name,
		Keywords: keywords,
		Action:   action,
	})
}

// AddRegexRule adds a rule that matches message content against regex patterns.
func (e *AutoModEngine) AddRegexRule(name string, patterns []string, action func(s *Session, msg *Message, reason string)) error {
	regexes := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return fmt.Errorf("invalid automod regex pattern %q: %w", p, err)
		}
		regexes = append(regexes, re)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.rules = append(e.rules, &LocalAutoModRule{
		Name:    name,
		Regexes: regexes,
		Action:  action,
	})
	return nil
}

// AddMentionLimitRule triggers when user mentions exceed limit.
func (e *AutoModEngine) AddMentionLimitRule(name string, maxMentions int, action func(s *Session, msg *Message, reason string)) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.rules = append(e.rules, &LocalAutoModRule{
		Name:        name,
		MaxMentions: maxMentions,
		Action:      action,
	})
}

// CheckMessage tests a message against all registered rules. Returns whether a match was found, rule name, and reason.
func (e *AutoModEngine) CheckMessage(msg *Message) (bool, string, string) {
	if msg == nil {
		return false, "", ""
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	lowerContent := strings.ToLower(msg.Content)

	for _, rule := range e.rules {
		// 1. Keyword check
		for _, kw := range rule.Keywords {
			if strings.Contains(lowerContent, strings.ToLower(kw)) {
				return true, rule.Name, fmt.Sprintf("contains blocked keyword %q", kw)
			}
		}

		// 2. Regex check
		for _, re := range rule.Regexes {
			if re.MatchString(msg.Content) {
				return true, rule.Name, fmt.Sprintf("matched regex rule %q", re.String())
			}
		}

		// 3. Mention count check
		if rule.MaxMentions > 0 && len(msg.Mentions) > rule.MaxMentions {
			return true, rule.Name, fmt.Sprintf("exceeded mention limit (%d > %d)", len(msg.Mentions), rule.MaxMentions)
		}
	}

	return false, "", ""
}

// Attach registers a MessageCreate handler on Session that automatically runs the rules.
func (e *AutoModEngine) Attach(s *Session) func() {
	return s.AddHandler(func(s *Session, ev *MessageCreate) {
		if ev == nil || ev.Message == nil {
			return
		}
		msg := ev.Message

		// Don't auto-mod bots or empty messages
		if msg.Author == nil || msg.Author.Bot {
			return
		}

		e.mu.RLock()
		defer e.mu.RUnlock()

		lowerContent := strings.ToLower(msg.Content)

		for _, rule := range e.rules {
			var matched bool
			var reason string

			for _, kw := range rule.Keywords {
				if strings.Contains(lowerContent, strings.ToLower(kw)) {
					matched = true
					reason = fmt.Sprintf("blocked keyword %q", kw)
					break
				}
			}

			if !matched {
				for _, re := range rule.Regexes {
					if re.MatchString(msg.Content) {
						matched = true
						reason = fmt.Sprintf("matched pattern %q", re.String())
						break
					}
				}
			}

			if !matched && rule.MaxMentions > 0 && len(msg.Mentions) > rule.MaxMentions {
				matched = true
				reason = fmt.Sprintf("mention spam (%d mentions)", len(msg.Mentions))
			}

			if matched && rule.Action != nil {
				go rule.Action(s, msg, reason)
			}
		}
	})
}
