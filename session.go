package micup

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"
)

// Session represents a high-level client session with Micup.
type Session struct {
	Token     string
	discovery *DiscoveryManager
	rest      *RESTClient
	gateway   *GatewayClient
	events    *EventHandlerRegistry
	options   *SessionOptions
}

// SessionOptions configures Session behavior.
type SessionOptions struct {
	BaseURL     string
	TokenPrefix string
	UserAgent   string
	HTTPClient  *http.Client
}

// Option modifies SessionOptions.
type Option func(*SessionOptions)

// WithBaseURL configures a custom platform base URL (default: "https://micup.gg").
func WithBaseURL(baseURL string) Option {
	return func(o *SessionOptions) {
		o.BaseURL = baseURL
	}
}

// WithTokenPrefix sets authorization header prefix (default: "Bot ").
func WithTokenPrefix(prefix string) Option {
	return func(o *SessionOptions) {
		o.TokenPrefix = prefix
	}
}

// WithUserAgent customizes the bot User-Agent string.
func WithUserAgent(ua string) Option {
	return func(o *SessionOptions) {
		o.UserAgent = ua
	}
}

// WithHTTPClient customizes the HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(o *SessionOptions) {
		o.HTTPClient = client
	}
}

// New creates a new Micup bot session.
func New(token string, opts ...Option) (*Session, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("bot token cannot be empty")
	}

	options := &SessionOptions{
		BaseURL:     DefaultBaseURL,
		TokenPrefix: "Bot ",
		UserAgent:   DefaultUserAgent,
		HTTPClient:  &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(options)
	}

	disc := NewDiscoveryManager(options.BaseURL, options.HTTPClient)
	rest := NewRESTClient(token, disc)
	rest.SetUserAgent(options.UserAgent)
	rest.SetTokenPrefix(options.TokenPrefix)

	s := &Session{
		Token:     token,
		discovery: disc,
		rest:      rest,
		events:    NewEventHandlerRegistry(),
		options:   options,
	}

	s.gateway = NewGatewayClient(s, token)
	return s, nil
}

// REST returns the low-level REST client.
func (s *Session) REST() *RESTClient {
	return s.rest
}

// Gateway returns the low-level Gateway client.
func (s *Session) Gateway() *GatewayClient {
	return s.gateway
}

// Open connects to the real-time Micup Gateway websocket.
func (s *Session) Open() error {
	return s.gateway.Open(context.Background())
}

// OpenContext connects with a caller-provided context.
func (s *Session) OpenContext(ctx context.Context) error {
	return s.gateway.Open(ctx)
}

// Close gracefully closes the gateway connection.
func (s *Session) Close() error {
	return s.gateway.Close()
}

// AddHandler registers an event handler function.
// Supported signatures:
//   - func(s *micup.Session, m *micup.MessageCreate)
//   - func(m *micup.MessageCreate)
//   - func(s *micup.Session, r *micup.Ready)
//   - func(s *micup.Session, g *micup.GuildCreate)
//   - etc.
func (s *Session) AddHandler(handler interface{}) func() {
	return s.events.Add(handler)
}

// UpdateStatus changes the bot presence status and activity.
func (s *Session) UpdateStatus(status string, activity *GatewayActivity) error {
	return s.gateway.UpdateStatus(status, activity)
}

// Convenience wrapper methods for REST actions

// UserMe fetches the authenticated bot user.
func (s *Session) UserMe() (*User, error) {
	return s.rest.UserMe(context.Background())
}

// GetUser fetches a user by ID.
func (s *Session) GetUser(userID string) (*User, error) {
	return s.rest.GetUser(context.Background(), userID)
}

// GetGuild fetches guild metadata.
func (s *Session) GetGuild(guildID string) (*Guild, error) {
	return s.rest.GetGuild(context.Background(), guildID)
}

// GetGuildChannels fetches channels for a guild.
func (s *Session) GetGuildChannels(guildID string) ([]*Channel, error) {
	return s.rest.GetGuildChannels(context.Background(), guildID)
}

// GetGuildMembers fetches guild members.
func (s *Session) GetGuildMembers(guildID string, limit int, after string) ([]*Member, error) {
	return s.rest.GetGuildMembers(context.Background(), guildID, limit, after)
}

// GetGuildMember fetches a single guild member.
func (s *Session) GetGuildMember(guildID, userID string) (*Member, error) {
	return s.rest.GetGuildMember(context.Background(), guildID, userID)
}

// GetGuildRoles fetches roles for a guild.
func (s *Session) GetGuildRoles(guildID string) ([]*Role, error) {
	return s.rest.GetGuildRoles(context.Background(), guildID)
}

// GetChannel fetches channel details.
func (s *Session) GetChannel(channelID string) (*Channel, error) {
	return s.rest.GetChannel(context.Background(), channelID)
}

// ChannelMessageSend sends a text message.
func (s *Session) ChannelMessageSend(channelID, content string) (*Message, error) {
	return s.rest.ChannelMessageSend(context.Background(), channelID, content)
}

// ChannelMessageSendEmbed sends an embed card.
func (s *Session) ChannelMessageSendEmbed(channelID string, embed *Embed) (*Message, error) {
	return s.rest.ChannelMessageSendEmbed(context.Background(), channelID, embed)
}

// ChannelMessageSendComplex sends a rich message.
func (s *Session) ChannelMessageSendComplex(channelID string, data *MessageSend) (*Message, error) {
	return s.rest.ChannelMessageSendComplex(context.Background(), channelID, data)
}

// ChannelMessageEdit edits an existing message.
func (s *Session) ChannelMessageEdit(channelID, messageID, content string) (*Message, error) {
	return s.rest.ChannelMessageEdit(context.Background(), channelID, messageID, content)
}

// ChannelMessageDelete deletes a message.
func (s *Session) ChannelMessageDelete(channelID, messageID string) error {
	return s.rest.ChannelMessageDelete(context.Background(), channelID, messageID)
}

// ChannelMessageAddReaction adds a reaction.
func (s *Session) ChannelMessageAddReaction(channelID, messageID, emoji string) error {
	return s.rest.ChannelMessageAddReaction(context.Background(), channelID, messageID, emoji)
}
