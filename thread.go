package micup

import (
	"context"
	"fmt"
)

// ThreadMetadata contains state information about a thread.
type ThreadMetadata struct {
	Archived            bool   `json:"archived"`
	AutoArchiveDuration int    `json:"auto_archive_duration"`
	ArchiveTimestamp    string `json:"archive_timestamp"`
	Locked              bool   `json:"locked"`
	Invitable           bool   `json:"invitable,omitempty"`
}

// ThreadMember represents a member who has joined a thread.
type ThreadMember struct {
	ID            string `json:"id,omitempty"`
	UserID        string `json:"user_id,omitempty"`
	JoinTimestamp string `json:"join_timestamp"`
	Flags         int    `json:"flags"`
}

// ThreadStartParams parameters for creating a new thread.
type ThreadStartParams struct {
	Name                string `json:"name"`
	AutoArchiveDuration int    `json:"auto_archive_duration,omitempty"` // 60, 1440, 4320, 10080 (minutes)
	Type                int    `json:"type,omitempty"`
	Invitable           bool   `json:"invitable,omitempty"`
}

// ChannelThreadStart creates a thread in a text channel without a starter message.
func (c *RESTClient) ChannelThreadStart(ctx context.Context, channelID string, params *ThreadStartParams) (*Channel, error) {
	var thread Channel
	err := c.RequestJSON(ctx, "POST", fmt.Sprintf("/channels/%s/threads", channelID), params, &thread)
	if err != nil {
		return nil, err
	}
	return &thread, nil
}

// MessageThreadStart creates a thread starting from an existing message.
func (c *RESTClient) MessageThreadStart(ctx context.Context, channelID, messageID string, params *ThreadStartParams) (*Channel, error) {
	var thread Channel
	err := c.RequestJSON(ctx, "POST", fmt.Sprintf("/channels/%s/messages/%s/threads", channelID, messageID), params, &thread)
	if err != nil {
		return nil, err
	}
	return &thread, nil
}

// ThreadListResponse response holding threads and thread members.
type ThreadListResponse struct {
	Threads []*Channel      `json:"threads"`
	Members []*ThreadMember `json:"members"`
	HasMore bool            `json:"has_more"`
}

// ChannelActiveThreads fetches active threads in a channel.
func (c *RESTClient) ChannelActiveThreads(ctx context.Context, channelID string) (*ThreadListResponse, error) {
	var res ThreadListResponse
	err := c.RequestJSON(ctx, "GET", fmt.Sprintf("/channels/%s/threads/active", channelID), nil, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// GuildActiveThreads fetches all active threads in a guild.
func (c *RESTClient) GuildActiveThreads(ctx context.Context, guildID string) (*ThreadListResponse, error) {
	var res ThreadListResponse
	err := c.RequestJSON(ctx, "GET", fmt.Sprintf("/guilds/%s/threads/active", guildID), nil, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Session convenience methods

// ChannelThreadStart creates a thread in a channel.
func (s *Session) ChannelThreadStart(channelID string, params *ThreadStartParams) (*Channel, error) {
	return s.rest.ChannelThreadStart(context.Background(), channelID, params)
}

// MessageThreadStart creates a thread attached to a message.
func (s *Session) MessageThreadStart(channelID, messageID string, params *ThreadStartParams) (*Channel, error) {
	return s.rest.MessageThreadStart(context.Background(), channelID, messageID, params)
}

// GuildActiveThreads lists all active threads in a guild.
func (s *Session) GuildActiveThreads(guildID string) (*ThreadListResponse, error) {
	return s.rest.GuildActiveThreads(context.Background(), guildID)
}
