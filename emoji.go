package micup

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// CreateEmojiParams defines fields required to upload a new guild custom emoji.
type CreateEmojiParams struct {
	Name  string   `json:"name"`
	Image string   `json:"image"` // Data URI, e.g. "data:image/png;base64,..."
	Roles []string `json:"roles,omitempty"`
}

// ModifyEmojiParams defines fields to update an existing custom emoji.
type ModifyEmojiParams struct {
	Name  *string   `json:"name,omitempty"`
	Roles *[]string `json:"roles,omitempty"`
}

// GuildEmojis returns all custom emojis for the given guild.
func (c *RESTClient) GuildEmojis(ctx context.Context, guildID string) ([]*Emoji, error) {
	var emojis []*Emoji
	err := c.RequestJSON(ctx, "GET", fmt.Sprintf("/guilds/%s/emojis", guildID), nil, &emojis)
	if err != nil {
		return nil, err
	}
	return emojis, nil
}

// GuildEmoji fetches a single custom emoji by ID.
func (c *RESTClient) GuildEmoji(ctx context.Context, guildID, emojiID string) (*Emoji, error) {
	var emoji Emoji
	err := c.RequestJSON(ctx, "GET", fmt.Sprintf("/guilds/%s/emojis/%s", guildID, emojiID), nil, &emoji)
	if err != nil {
		return nil, err
	}
	return &emoji, nil
}

// GuildEmojiCreate uploads a new custom emoji to a guild.
func (c *RESTClient) GuildEmojiCreate(ctx context.Context, guildID string, params *CreateEmojiParams) (*Emoji, error) {
	var created Emoji
	err := c.RequestJSON(ctx, "POST", fmt.Sprintf("/guilds/%s/emojis", guildID), params, &created)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

// GuildEmojiModify modifies an existing guild emoji's name or allowed roles.
func (c *RESTClient) GuildEmojiModify(ctx context.Context, guildID, emojiID string, params *ModifyEmojiParams) (*Emoji, error) {
	var modified Emoji
	err := c.RequestJSON(ctx, "PATCH", fmt.Sprintf("/guilds/%s/emojis/%s", guildID, emojiID), params, &modified)
	if err != nil {
		return nil, err
	}
	return &modified, nil
}

// GuildEmojiDelete deletes an emoji from a guild.
func (c *RESTClient) GuildEmojiDelete(ctx context.Context, guildID, emojiID string) error {
	return c.RequestJSON(ctx, "DELETE", fmt.Sprintf("/guilds/%s/emojis/%s", guildID, emojiID), nil, nil)
}

// Reaction endpoints

// MessageReactionAdd adds a reaction (emoji or unicode) to a message.
func (c *RESTClient) MessageReactionAdd(ctx context.Context, channelID, messageID, emoji string) error {
	encoded := url.PathEscape(emoji)
	return c.RequestJSON(ctx, "PUT", fmt.Sprintf("/channels/%s/messages/%s/reactions/%s/@me", channelID, messageID, encoded), nil, nil)
}

// MessageReactionRemove removes a reaction from a message. If userID is empty or "@me", removes own reaction.
func (c *RESTClient) MessageReactionRemove(ctx context.Context, channelID, messageID, emoji, userID string) error {
	if userID == "" {
		userID = "@me"
	}
	encoded := url.PathEscape(emoji)
	return c.RequestJSON(ctx, "DELETE", fmt.Sprintf("/channels/%s/messages/%s/reactions/%s/%s", channelID, messageID, encoded, userID), nil, nil)
}

// MessageReactionsRemoveAll removes all reactions from a message.
func (c *RESTClient) MessageReactionsRemoveAll(ctx context.Context, channelID, messageID string) error {
	return c.RequestJSON(ctx, "DELETE", fmt.Sprintf("/channels/%s/messages/%s/reactions", channelID, messageID), nil, nil)
}

// MessageReactionsRemoveEmoji removes all reactions of a specific emoji from a message.
func (c *RESTClient) MessageReactionsRemoveEmoji(ctx context.Context, channelID, messageID, emoji string) error {
	encoded := url.PathEscape(emoji)
	return c.RequestJSON(ctx, "DELETE", fmt.Sprintf("/channels/%s/messages/%s/reactions/%s", channelID, messageID, encoded), nil, nil)
}

// MessageReactions fetches a list of users that reacted with a specific emoji.
func (c *RESTClient) MessageReactions(ctx context.Context, channelID, messageID, emoji string, limit int, after string) ([]*User, error) {
	encoded := url.PathEscape(emoji)
	endpoint := fmt.Sprintf("/channels/%s/messages/%s/reactions/%s", channelID, messageID, encoded)

	values := url.Values{}
	if limit > 0 {
		values.Set("limit", strconv.Itoa(limit))
	}
	if after != "" {
		values.Set("after", after)
	}
	if len(values) > 0 {
		endpoint = fmt.Sprintf("%s?%s", endpoint, values.Encode())
	}

	var users []*User
	err := c.RequestJSON(ctx, "GET", endpoint, nil, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// Session convenience methods

// GuildEmojis lists emojis in a guild.
func (s *Session) GuildEmojis(guildID string) ([]*Emoji, error) {
	return s.rest.GuildEmojis(context.Background(), guildID)
}

// GuildEmoji fetches a guild emoji.
func (s *Session) GuildEmoji(guildID, emojiID string) (*Emoji, error) {
	return s.rest.GuildEmoji(context.Background(), guildID, emojiID)
}

// GuildEmojiCreate creates a new emoji in a guild.
func (s *Session) GuildEmojiCreate(guildID string, params *CreateEmojiParams) (*Emoji, error) {
	return s.rest.GuildEmojiCreate(context.Background(), guildID, params)
}

// GuildEmojiModify updates a guild emoji.
func (s *Session) GuildEmojiModify(guildID, emojiID string, params *ModifyEmojiParams) (*Emoji, error) {
	return s.rest.GuildEmojiModify(context.Background(), guildID, emojiID, params)
}

// GuildEmojiDelete deletes an emoji.
func (s *Session) GuildEmojiDelete(guildID, emojiID string) error {
	return s.rest.GuildEmojiDelete(context.Background(), guildID, emojiID)
}

// MessageReactionAdd adds a reaction to a message.
func (s *Session) MessageReactionAdd(channelID, messageID, emoji string) error {
	return s.rest.MessageReactionAdd(context.Background(), channelID, messageID, emoji)
}

// MessageReactionRemove removes a reaction from a message.
func (s *Session) MessageReactionRemove(channelID, messageID, emoji string, userID ...string) error {
	target := "@me"
	if len(userID) > 0 && userID[0] != "" {
		target = userID[0]
	}
	return s.rest.MessageReactionRemove(context.Background(), channelID, messageID, emoji, target)
}

// MessageReactionsRemoveAll removes all reactions.
func (s *Session) MessageReactionsRemoveAll(channelID, messageID string) error {
	return s.rest.MessageReactionsRemoveAll(context.Background(), channelID, messageID)
}

// MessageReactionsRemoveEmoji removes all reactions of an emoji.
func (s *Session) MessageReactionsRemoveEmoji(channelID, messageID, emoji string) error {
	return s.rest.MessageReactionsRemoveEmoji(context.Background(), channelID, messageID, emoji)
}

// MessageReactions lists users who reacted with an emoji.
func (s *Session) MessageReactions(channelID, messageID, emoji string, limit int) ([]*User, error) {
	return s.rest.MessageReactions(context.Background(), channelID, messageID, emoji, limit, "")
}
