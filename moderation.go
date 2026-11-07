package micup

import (
	"context"
	"fmt"
)

// GuildBan represents a ban entry in a guild.
type GuildBan struct {
	User   *User  `json:"user"`
	Reason string `json:"reason,omitempty"`
}

// AuditLog represents guild audit logs.
type AuditLog struct {
	Webhooks []*Webhook        `json:"webhooks"`
	Users    []*User           `json:"users"`
	Entries  []*AuditLogEntry  `json:"audit_log_entries"`
}

// AuditLogEntry represents a single audit log record.
type AuditLogEntry struct {
	TargetID   string `json:"target_id"`
	UserID     string `json:"user_id"`
	ID         string `json:"id"`
	ActionType int    `json:"action_type"`
	Reason     string `json:"reason,omitempty"`
}

// ChannelMessagesBulkDelete deletes multiple messages in a single request (up to 100).
func (c *RESTClient) ChannelMessagesBulkDelete(ctx context.Context, channelID string, messageIDs []string) error {
	body := map[string][]string{
		"messages": messageIDs,
	}
	return c.RequestJSON(ctx, "POST", fmt.Sprintf("/channels/%s/messages/bulk-delete", channelID), body, nil)
}

// ChannelTyping sends a typing indicator to a channel.
func (c *RESTClient) ChannelTyping(ctx context.Context, channelID string) error {
	return c.RequestJSON(ctx, "POST", fmt.Sprintf("/channels/%s/typing", channelID), nil, nil)
}

// GuildBans fetches all bans for a guild.
func (c *RESTClient) GuildBans(ctx context.Context, guildID string) ([]*GuildBan, error) {
	var bans []*GuildBan
	err := c.RequestJSON(ctx, "GET", fmt.Sprintf("/guilds/%s/bans", guildID), nil, &bans)
	if err != nil {
		return nil, err
	}
	return bans, nil
}

// GuildAuditLogs fetches audit log entries for a guild.
func (c *RESTClient) GuildAuditLogs(ctx context.Context, guildID string) (*AuditLog, error) {
	var logs AuditLog
	err := c.RequestJSON(ctx, "GET", fmt.Sprintf("/guilds/%s/audit-logs", guildID), nil, &logs)
	if err != nil {
		return nil, err
	}
	return &logs, nil
}

// GuildPruneCount checks how many inactive members would be removed.
func (c *RESTClient) GuildPruneCount(ctx context.Context, guildID string, days int) (int, error) {
	endpoint := fmt.Sprintf("/guilds/%s/prune?days=%d", guildID, days)
	var resp struct {
		Pruned int `json:"pruned"`
	}
	err := c.RequestJSON(ctx, "GET", endpoint, nil, &resp)
	if err != nil {
		return 0, err
	}
	return resp.Pruned, nil
}

// GuildPrune executes member pruning.
func (c *RESTClient) GuildPrune(ctx context.Context, guildID string, days int) (int, error) {
	endpoint := fmt.Sprintf("/guilds/%s/prune?days=%d", guildID, days)
	var resp struct {
		Pruned int `json:"pruned"`
	}
	err := c.RequestJSON(ctx, "POST", endpoint, nil, &resp)
	if err != nil {
		return 0, err
	}
	return resp.Pruned, nil
}

// Session convenience methods

// ChannelMessagesBulkDelete deletes messages in bulk.
func (s *Session) ChannelMessagesBulkDelete(channelID string, messageIDs []string) error {
	return s.rest.ChannelMessagesBulkDelete(context.Background(), channelID, messageIDs)
}

// ChannelTyping triggers the typing indicator.
func (s *Session) ChannelTyping(channelID string) error {
	return s.rest.ChannelTyping(context.Background(), channelID)
}

// GuildBans lists all bans in a guild.
func (s *Session) GuildBans(guildID string) ([]*GuildBan, error) {
	return s.rest.GuildBans(context.Background(), guildID)
}

// GuildAuditLogs fetches audit logs.
func (s *Session) GuildAuditLogs(guildID string) (*AuditLog, error) {
	return s.rest.GuildAuditLogs(context.Background(), guildID)
}
