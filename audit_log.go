package micup

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// Audit Log action type constants.
const (
	AuditLogActionGuildUpdate                          = 1
	AuditLogActionChannelCreate                        = 10
	AuditLogActionChannelUpdate                        = 11
	AuditLogActionChannelDelete                        = 12
	AuditLogActionChannelOverwriteCreate               = 13
	AuditLogActionChannelOverwriteUpdate               = 14
	AuditLogActionChannelOverwriteDelete               = 15
	AuditLogActionMemberKick                           = 20
	AuditLogActionMemberPrune                          = 21
	AuditLogActionMemberBanAdd                         = 22
	AuditLogActionMemberBanRemove                      = 23
	AuditLogActionMemberUpdate                         = 24
	AuditLogActionMemberRoleUpdate                     = 25
	AuditLogActionMemberMove                           = 26
	AuditLogActionMemberDisconnect                     = 27
	AuditLogActionBotAdd                               = 28
	AuditLogActionRoleCreate                           = 30
	AuditLogActionRoleUpdate                           = 31
	AuditLogActionRoleDelete                           = 32
	AuditLogActionInviteCreate                         = 40
	AuditLogActionInviteUpdate                         = 41
	AuditLogActionInviteDelete                         = 42
	AuditLogActionWebhookCreate                        = 50
	AuditLogActionWebhookUpdate                        = 51
	AuditLogActionWebhookDelete                        = 52
	AuditLogActionEmojiCreate                          = 60
	AuditLogActionEmojiUpdate                          = 61
	AuditLogActionEmojiDelete                          = 62
	AuditLogActionMessageDelete                        = 72
	AuditLogActionMessageBulkDelete                    = 73
	AuditLogActionMessagePin                           = 74
	AuditLogActionMessageUnpin                         = 75
	AuditLogActionIntegrationCreate                    = 80
	AuditLogActionIntegrationUpdate                    = 81
	AuditLogActionIntegrationDelete                    = 82
	AuditLogActionStageInstanceCreate                  = 83
	AuditLogActionStageInstanceUpdate                  = 84
	AuditLogActionStageInstanceDelete                  = 85
	AuditLogActionStickerCreate                        = 90
	AuditLogActionStickerUpdate                        = 91
	AuditLogActionStickerDelete                        = 92
	AuditLogActionGuildScheduledEventCreate            = 100
	AuditLogActionGuildScheduledEventUpdate            = 101
	AuditLogActionGuildScheduledEventDelete            = 102
	AuditLogActionThreadCreate                         = 110
	AuditLogActionThreadUpdate                         = 111
	AuditLogActionThreadDelete                         = 112
	AuditLogActionAutoModerationRuleCreate             = 140
	AuditLogActionAutoModerationRuleUpdate             = 141
	AuditLogActionAutoModerationRuleDelete             = 142
	AuditLogActionAutoModerationBlockMessage           = 143
	AuditLogActionAutoModerationFlagToChannel          = 144
	AuditLogActionAutoModerationUserCommDisabled       = 145
)

// AuditLogChange represents a diff in an audit log entry.
type AuditLogChange struct {
	Key      string      `json:"key"`
	NewValue interface{} `json:"new_value,omitempty"`
	OldValue interface{} `json:"old_value,omitempty"`
}

// AuditLogOptionsData represents extra optional information for certain audit log action types.
type AuditLogOptionsData struct {
	ApplicationID                 string `json:"application_id,omitempty"`
	AutoModerationRuleName        string `json:"auto_moderation_rule_name,omitempty"`
	AutoModerationRuleTriggerType string `json:"auto_moderation_rule_trigger_type,omitempty"`
	ChannelID                     string `json:"channel_id,omitempty"`
	Count                         string `json:"count,omitempty"`
	DeleteMemberDays              string `json:"delete_member_days,omitempty"`
	ID                            string `json:"id,omitempty"`
	MembersRemoved                string `json:"members_removed,omitempty"`
	MessageID                     string `json:"message_id,omitempty"`
	RoleName                      string `json:"role_name,omitempty"`
	Type                          string `json:"type,omitempty"`
	Status                        string `json:"status,omitempty"`
}

// AuditLogEntry represents a single audit log event record.
type AuditLogEntry struct {
	ID         string               `json:"id"`
	TargetID   string               `json:"target_id"`
	UserID     string               `json:"user_id"`
	ActionType int                  `json:"action_type"`
	Changes    []*AuditLogChange    `json:"changes,omitempty"`
	Options    *AuditLogOptionsData `json:"options,omitempty"`
	Reason     string               `json:"reason,omitempty"`
}

// AuditLog contains a list of audit log entries along with resolved users and webhooks.
type AuditLog struct {
	Webhooks []*Webhook       `json:"webhooks"`
	Users    []*User          `json:"users"`
	Entries  []*AuditLogEntry `json:"audit_log_entries"`
	Threads  []*Channel       `json:"threads,omitempty"`
}

// AuditLogQuery specifies filtering criteria for fetching audit logs.
type AuditLogQuery struct {
	UserID     string
	ActionType int
	Before     string
	After      string
	Limit      int
}

// GuildAuditLogs fetches audit log entries for a guild with optional query filters.
func (c *RESTClient) GuildAuditLogs(ctx context.Context, guildID string, query *AuditLogQuery) (*AuditLog, error) {
	endpoint := fmt.Sprintf("/guilds/%s/audit-logs", guildID)
	if query != nil {
		params := url.Values{}
		if query.UserID != "" {
			params.Set("user_id", query.UserID)
		}
		if query.ActionType > 0 {
			params.Set("action_type", strconv.Itoa(query.ActionType))
		}
		if query.Before != "" {
			params.Set("before", query.Before)
		}
		if query.After != "" {
			params.Set("after", query.After)
		}
		if query.Limit > 0 {
			params.Set("limit", strconv.Itoa(query.Limit))
		}
		if len(params) > 0 {
			endpoint = fmt.Sprintf("%s?%s", endpoint, params.Encode())
		}
	}

	var logs AuditLog
	err := c.RequestJSON(ctx, "GET", endpoint, nil, &logs)
	if err != nil {
		return nil, err
	}
	return &logs, nil
}

// Session convenience methods

// GuildAuditLogs fetches audit log records for a guild.
func (s *Session) GuildAuditLogs(guildID string, query ...*AuditLogQuery) (*AuditLog, error) {
	var q *AuditLogQuery
	if len(query) > 0 {
		q = query[0]
	}
	return s.rest.GuildAuditLogs(context.Background(), guildID, q)
}
