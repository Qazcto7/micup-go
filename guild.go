package micup

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// GetGuild fetches guild metadata by guild ID.
func (c *RESTClient) GetGuild(ctx context.Context, guildID string) (*Guild, error) {
	var guild Guild
	err := c.RequestJSON(ctx, "GET", fmt.Sprintf("/guilds/%s", guildID), nil, &guild)
	if err != nil {
		return nil, err
	}
	return &guild, nil
}

// GetGuildChannels fetches all channels in a guild.
func (c *RESTClient) GetGuildChannels(ctx context.Context, guildID string) ([]*Channel, error) {
	var channels []*Channel
	err := c.RequestJSON(ctx, "GET", fmt.Sprintf("/guilds/%s/channels", guildID), nil, &channels)
	if err != nil {
		return nil, err
	}
	return channels, nil
}

// ChannelParams defines parameters for creating or modifying a channel.
type ChannelParams struct {
	Name                 string                 `json:"name"`
	Type                 int                    `json:"type"`
	Topic                string                 `json:"topic,omitempty"`
	Bitrate              int                    `json:"bitrate,omitempty"`
	UserLimit            int                    `json:"user_limit,omitempty"`
	RateLimitPerUser     int                    `json:"rate_limit_per_user,omitempty"`
	Position             int                    `json:"position,omitempty"`
	PermissionOverwrites []*PermissionOverwrite `json:"permission_overwrites,omitempty"`
	ParentID             string                 `json:"parent_id,omitempty"`
	NSFW                 bool                   `json:"nsfw,omitempty"`
}

// CreateGuildChannel creates a channel in the guild.
func (c *RESTClient) CreateGuildChannel(ctx context.Context, guildID string, data *ChannelParams) (*Channel, error) {
	var channel Channel
	err := c.RequestJSON(ctx, "POST", fmt.Sprintf("/guilds/%s/channels", guildID), data, &channel)
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

// GetGuildMembers fetches members in a guild with pagination.
func (c *RESTClient) GetGuildMembers(ctx context.Context, guildID string, limit int, after string) ([]*Member, error) {
	values := url.Values{}
	if limit > 0 {
		values.Set("limit", strconv.Itoa(limit))
	}
	if after != "" {
		values.Set("after", after)
	}

	endpoint := fmt.Sprintf("/guilds/%s/members", guildID)
	if len(values) > 0 {
		endpoint += "?" + values.Encode()
	}

	var members []*Member
	err := c.RequestJSON(ctx, "GET", endpoint, nil, &members)
	if err != nil {
		return nil, err
	}
	return members, nil
}

// GetGuildMember fetches a specific member in a guild.
func (c *RESTClient) GetGuildMember(ctx context.Context, guildID, userID string) (*Member, error) {
	var member Member
	err := c.RequestJSON(ctx, "GET", fmt.Sprintf("/guilds/%s/members/%s", guildID, userID), nil, &member)
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// GetGuildRoles fetches all roles in a guild.
func (c *RESTClient) GetGuildRoles(ctx context.Context, guildID string) ([]*Role, error) {
	var roles []*Role
	err := c.RequestJSON(ctx, "GET", fmt.Sprintf("/guilds/%s/roles", guildID), nil, &roles)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// RoleParams parameters for creating or modifying a role.
type RoleParams struct {
	Name        string `json:"name,omitempty"`
	Permissions string `json:"permissions,omitempty"`
	Color       int    `json:"color,omitempty"`
	Hoist       *bool  `json:"hoist,omitempty"`
	Mentionable *bool  `json:"mentionable,omitempty"`
}

// CreateGuildRole creates a role in the specified guild.
func (c *RESTClient) CreateGuildRole(ctx context.Context, guildID string, data *RoleParams) (*Role, error) {
	var role Role
	err := c.RequestJSON(ctx, "POST", fmt.Sprintf("/guilds/%s/roles", guildID), data, &role)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// ModifyGuildRole updates an existing role.
func (c *RESTClient) ModifyGuildRole(ctx context.Context, guildID, roleID string, data *RoleParams) (*Role, error) {
	var role Role
	err := c.RequestJSON(ctx, "PATCH", fmt.Sprintf("/guilds/%s/roles/%s", guildID, roleID), data, &role)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// DeleteGuildRole deletes a role from the guild.
func (c *RESTClient) DeleteGuildRole(ctx context.Context, guildID, roleID string) error {
	return c.RequestJSON(ctx, "DELETE", fmt.Sprintf("/guilds/%s/roles/%s", guildID, roleID), nil, nil)
}

// GuildMemberAddRole assigns a role to a member.
func (c *RESTClient) GuildMemberAddRole(ctx context.Context, guildID, userID, roleID string) error {
	return c.RequestJSON(ctx, "PUT", fmt.Sprintf("/guilds/%s/members/%s/roles/%s", guildID, userID, roleID), nil, nil)
}

// GuildMemberRemoveRole removes a role from a member.
func (c *RESTClient) GuildMemberRemoveRole(ctx context.Context, guildID, userID, roleID string) error {
	return c.RequestJSON(ctx, "DELETE", fmt.Sprintf("/guilds/%s/members/%s/roles/%s", guildID, userID, roleID), nil, nil)
}

// GuildMemberKick kicks a member from the guild.
func (c *RESTClient) GuildMemberKick(ctx context.Context, guildID, userID string) error {
	return c.RequestJSON(ctx, "DELETE", fmt.Sprintf("/guilds/%s/members/%s", guildID, userID), nil, nil)
}

// GuildMemberBan bans a member from the guild.
func (c *RESTClient) GuildMemberBan(ctx context.Context, guildID, userID string, deleteMessageDays int, reason string) error {
	body := map[string]interface{}{}
	if deleteMessageDays > 0 {
		body["delete_message_days"] = deleteMessageDays
	}
	if reason != "" {
		body["reason"] = reason
	}
	return c.RequestJSON(ctx, "PUT", fmt.Sprintf("/guilds/%s/bans/%s", guildID, userID), body, nil)
}

// GuildMemberUnban unbans a user from the guild.
func (c *RESTClient) GuildMemberUnban(ctx context.Context, guildID, userID string) error {
	return c.RequestJSON(ctx, "DELETE", fmt.Sprintf("/guilds/%s/bans/%s", guildID, userID), nil, nil)
}

// Session convenience methods

// GuildMemberKick kicks a member from the guild.
func (s *Session) GuildMemberKick(guildID, userID string) error {
	return s.rest.GuildMemberKick(context.Background(), guildID, userID)
}

// GuildMemberBan bans a member from the guild.
func (s *Session) GuildMemberBan(guildID, userID string, deleteMessageDays int, reason string) error {
	return s.rest.GuildMemberBan(context.Background(), guildID, userID, deleteMessageDays, reason)
}

// GuildMemberUnban unbans a user from the guild.
func (s *Session) GuildMemberUnban(guildID, userID string) error {
	return s.rest.GuildMemberUnban(context.Background(), guildID, userID)
}

// GuildChannelCreate creates a channel in a guild.
func (s *Session) GuildChannelCreate(guildID string, params *ChannelParams) (*Channel, error) {
	return s.rest.CreateGuildChannel(context.Background(), guildID, params)
}

// ChannelDelete deletes a channel.
func (s *Session) ChannelDelete(channelID string) error {
	return s.rest.DeleteChannel(context.Background(), channelID)
}
