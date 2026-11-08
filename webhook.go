package micup

import (
	"context"
	"fmt"
	"net/http"
)

// WebhookType values.
const (
	WebhookTypeIncoming        = 1
	WebhookTypeChannelFollower = 2
	WebhookTypeApplication     = 3
)

// WebhookParams defines parameters for creating or modifying a webhook.
type WebhookParams struct {
	Name      string `json:"name,omitempty"`
	Avatar    string `json:"avatar,omitempty"`
	ChannelID string `json:"channel_id,omitempty"`
}

// WebhookExecuteParams defines payload parameters for executing a webhook.
type WebhookExecuteParams struct {
	Content    string      `json:"content,omitempty"`
	Username   string      `json:"username,omitempty"`
	AvatarURL  string      `json:"avatar_url,omitempty"`
	TTS        bool        `json:"tts,omitempty"`
	Embeds     []*Embed    `json:"embeds,omitempty"`
	Components []Component `json:"components,omitempty"`
}

// CreateWebhook creates a new webhook in a channel.
func (c *RESTClient) CreateWebhook(ctx context.Context, channelID string, params *WebhookParams) (*Webhook, error) {
	endpoint := fmt.Sprintf("/channels/%s/webhooks", channelID)
	var wh Webhook
	if err := c.RequestJSON(ctx, http.MethodPost, endpoint, params, &wh); err != nil {
		return nil, err
	}
	return &wh, nil
}

// GetChannelWebhooks returns all webhooks in a channel.
func (c *RESTClient) GetChannelWebhooks(ctx context.Context, channelID string) ([]*Webhook, error) {
	endpoint := fmt.Sprintf("/channels/%s/webhooks", channelID)
	var list []*Webhook
	if err := c.RequestJSON(ctx, http.MethodGet, endpoint, nil, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// GetGuildWebhooks returns all webhooks in a guild.
func (c *RESTClient) GetGuildWebhooks(ctx context.Context, guildID string) ([]*Webhook, error) {
	endpoint := fmt.Sprintf("/guilds/%s/webhooks", guildID)
	var list []*Webhook
	if err := c.RequestJSON(ctx, http.MethodGet, endpoint, nil, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// GetWebhook returns a webhook by its ID.
func (c *RESTClient) GetWebhook(ctx context.Context, webhookID string) (*Webhook, error) {
	endpoint := fmt.Sprintf("/webhooks/%s", webhookID)
	var wh Webhook
	if err := c.RequestJSON(ctx, http.MethodGet, endpoint, nil, &wh); err != nil {
		return nil, err
	}
	return &wh, nil
}

// GetWebhookWithToken returns a webhook using its token (no bot auth required).
func (c *RESTClient) GetWebhookWithToken(ctx context.Context, webhookID, token string) (*Webhook, error) {
	endpoint := fmt.Sprintf("/webhooks/%s/%s", webhookID, token)
	var wh Webhook
	if err := c.RequestJSON(ctx, http.MethodGet, endpoint, nil, &wh); err != nil {
		return nil, err
	}
	return &wh, nil
}

// ModifyWebhook edits an existing webhook.
func (c *RESTClient) ModifyWebhook(ctx context.Context, webhookID string, params *WebhookParams) (*Webhook, error) {
	endpoint := fmt.Sprintf("/webhooks/%s", webhookID)
	var wh Webhook
	if err := c.RequestJSON(ctx, http.MethodPatch, endpoint, params, &wh); err != nil {
		return nil, err
	}
	return &wh, nil
}

// DeleteWebhook removes a webhook.
func (c *RESTClient) DeleteWebhook(ctx context.Context, webhookID string) error {
	endpoint := fmt.Sprintf("/webhooks/%s", webhookID)
	resp, err := c.Request(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// DeleteWebhookWithToken removes a webhook using its token.
func (c *RESTClient) DeleteWebhookWithToken(ctx context.Context, webhookID, token string) error {
	endpoint := fmt.Sprintf("/webhooks/%s/%s", webhookID, token)
	resp, err := c.Request(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// ExecuteWebhook posts a message through a webhook.
func (c *RESTClient) ExecuteWebhook(ctx context.Context, webhookID, token string, wait bool, params *WebhookExecuteParams) (*Message, error) {
	endpoint := fmt.Sprintf("/webhooks/%s/%s", webhookID, token)
	if wait {
		endpoint += "?wait=true"
		var msg Message
		if err := c.RequestJSON(ctx, http.MethodPost, endpoint, params, &msg); err != nil {
			return nil, err
		}
		return &msg, nil
	}

	resp, err := c.Request(ctx, http.MethodPost, endpoint, params)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	return nil, nil
}

// Session convenience methods

// WebhookCreate creates a webhook in the specified channel.
func (s *Session) WebhookCreate(channelID, name string) (*Webhook, error) {
	return s.rest.CreateWebhook(context.Background(), channelID, &WebhookParams{Name: name})
}

// WebhookGet fetches a webhook by ID.
func (s *Session) WebhookGet(webhookID string) (*Webhook, error) {
	return s.rest.GetWebhook(context.Background(), webhookID)
}

// WebhookExecute executes a webhook.
func (s *Session) WebhookExecute(webhookID, token string, wait bool, params *WebhookExecuteParams) (*Message, error) {
	return s.rest.ExecuteWebhook(context.Background(), webhookID, token, wait, params)
}
