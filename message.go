package micup

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// GetChannel fetches a channel by its ID.
func (c *RESTClient) GetChannel(ctx context.Context, channelID string) (*Channel, error) {
	var channel Channel
	err := c.RequestJSON(ctx, "GET", fmt.Sprintf("/channels/%s", channelID), nil, &channel)
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

// ModifyChannel updates channel settings.
func (c *RESTClient) ModifyChannel(ctx context.Context, channelID string, data *ChannelParams) (*Channel, error) {
	var channel Channel
	err := c.RequestJSON(ctx, "PATCH", fmt.Sprintf("/channels/%s", channelID), data, &channel)
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

// DeleteChannel deletes a channel.
func (c *RESTClient) DeleteChannel(ctx context.Context, channelID string) error {
	return c.RequestJSON(ctx, "DELETE", fmt.Sprintf("/channels/%s", channelID), nil, nil)
}

// ChannelMessageSend sends a basic text message to a channel.
func (c *RESTClient) ChannelMessageSend(ctx context.Context, channelID, content string) (*Message, error) {
	return c.ChannelMessageSendComplex(ctx, channelID, &MessageSend{
		Content: content,
	})
}

// ChannelMessageSendEmbed sends a message containing an embed.
func (c *RESTClient) ChannelMessageSendEmbed(ctx context.Context, channelID string, embed *Embed) (*Message, error) {
	return c.ChannelMessageSendComplex(ctx, channelID, &MessageSend{
		Embeds: []*Embed{embed},
	})
}

// ChannelMessageSendComplex sends a complex message with embeds, tts, reply, etc.
func (c *RESTClient) ChannelMessageSendComplex(ctx context.Context, channelID string, data *MessageSend) (*Message, error) {
	var msg Message
	err := c.RequestJSON(ctx, "POST", fmt.Sprintf("/channels/%s/messages", channelID), data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// ChannelMessageEdit edits an existing message.
func (c *RESTClient) ChannelMessageEdit(ctx context.Context, channelID, messageID, content string) (*Message, error) {
	var msg Message
	data := &MessageSend{Content: content}
	err := c.RequestJSON(ctx, "PATCH", fmt.Sprintf("/channels/%s/messages/%s", channelID, messageID), data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// ChannelMessageDelete deletes a message from a channel.
func (c *RESTClient) ChannelMessageDelete(ctx context.Context, channelID, messageID string) error {
	return c.RequestJSON(ctx, "DELETE", fmt.Sprintf("/channels/%s/messages/%s", channelID, messageID), nil, nil)
}

// ChannelMessages fetches message history for a channel.
func (c *RESTClient) ChannelMessages(ctx context.Context, channelID string, limit int, around, before, after string) ([]*Message, error) {
	values := url.Values{}
	if limit > 0 {
		values.Set("limit", strconv.Itoa(limit))
	}
	if around != "" {
		values.Set("around", around)
	}
	if before != "" {
		values.Set("before", before)
	}
	if after != "" {
		values.Set("after", after)
	}

	endpoint := fmt.Sprintf("/channels/%s/messages", channelID)
	if len(values) > 0 {
		endpoint += "?" + values.Encode()
	}

	var messages []*Message
	err := c.RequestJSON(ctx, "GET", endpoint, nil, &messages)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// ChannelMessagePin pins a message in a channel.
func (c *RESTClient) ChannelMessagePin(ctx context.Context, channelID, messageID string) error {
	return c.RequestJSON(ctx, "PUT", fmt.Sprintf("/channels/%s/pins/%s", channelID, messageID), nil, nil)
}

// ChannelMessageUnpin unpins a message.
func (c *RESTClient) ChannelMessageUnpin(ctx context.Context, channelID, messageID string) error {
	return c.RequestJSON(ctx, "DELETE", fmt.Sprintf("/channels/%s/pins/%s", channelID, messageID), nil, nil)
}

// ChannelPins fetches all pinned messages in a channel.
func (c *RESTClient) ChannelPins(ctx context.Context, channelID string) ([]*Message, error) {
	var pins []*Message
	err := c.RequestJSON(ctx, "GET", fmt.Sprintf("/channels/%s/pins", channelID), nil, &pins)
	if err != nil {
		return nil, err
	}
	return pins, nil
}

// ChannelMessageAddReaction adds a reaction to a message.
func (c *RESTClient) ChannelMessageAddReaction(ctx context.Context, channelID, messageID, emoji string) error {
	encodedEmoji := url.PathEscape(emoji)
	return c.RequestJSON(ctx, "PUT", fmt.Sprintf("/channels/%s/messages/%s/reactions/%s/@me", channelID, messageID, encodedEmoji), nil, nil)
}

// ChannelMessageRemoveReaction removes a user's reaction from a message.
func (c *RESTClient) ChannelMessageRemoveReaction(ctx context.Context, channelID, messageID, emoji string) error {
	encodedEmoji := url.PathEscape(emoji)
	return c.RequestJSON(ctx, "DELETE", fmt.Sprintf("/channels/%s/messages/%s/reactions/%s/@me", channelID, messageID, encodedEmoji), nil, nil)
}

// GetInvite fetches invite information by invite code.
func (c *RESTClient) GetInvite(ctx context.Context, code string) (*Invite, error) {
	var invite Invite
	err := c.RequestJSON(ctx, "GET", fmt.Sprintf("/invites/%s", code), nil, &invite)
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

// DeleteInvite deletes an invite.
func (c *RESTClient) DeleteInvite(ctx context.Context, code string) error {
	return c.RequestJSON(ctx, "DELETE", fmt.Sprintf("/invites/%s", code), nil, nil)
}
