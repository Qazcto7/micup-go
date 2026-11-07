package micup

import (
	"encoding/json"
	"time"
)

// ChannelType values.
const (
	ChannelTypeGuildText          = 0
	ChannelTypeDM                 = 1
	ChannelTypeGuildVoice         = 2
	ChannelTypeGroupDM            = 3
	ChannelTypeGuildCategory      = 4
	ChannelTypeGuildNews          = 5
	ChannelTypeGuildNewsThread    = 10
	ChannelTypeGuildPublicThread  = 11
	ChannelTypeGuildPrivateThread = 12
	ChannelTypeGuildStageVoice    = 13
	ChannelTypeGuildForum         = 15
)

// User represents a user account or bot in Micup.
type User struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
	Bot           bool   `json:"bot,omitempty"`
	System        bool   `json:"system,omitempty"`
	MFAEnabled    bool   `json:"mfa_enabled,omitempty"`
	Banner        string `json:"banner,omitempty"`
	AccentColor   int    `json:"accent_color,omitempty"`
	Locale        string `json:"locale,omitempty"`
	Verified      bool   `json:"verified,omitempty"`
	Email         string `json:"email,omitempty"`
	Flags         int    `json:"flags,omitempty"`
	PremiumType   int    `json:"premium_type,omitempty"`
	PublicFlags   int    `json:"public_flags,omitempty"`
}

// Mention returns the string format to mention this user in a message.
func (u *User) Mention() string {
	return "<@" + u.ID + ">"
}

// Guild represents a server (community) in Micup.
type Guild struct {
	ID                          string        `json:"id"`
	Name                        string        `json:"name"`
	Icon                        string        `json:"icon"`
	Description                 string        `json:"description,omitempty"`
	Splash                      string        `json:"splash,omitempty"`
	DiscoverySplash             string        `json:"discovery_splash,omitempty"`
	OwnerID                     string        `json:"owner_id"`
	Region                      string        `json:"region,omitempty"`
	AFKChannelID                string        `json:"afk_channel_id,omitempty"`
	AFKTimeout                  int           `json:"afk_timeout,omitempty"`
	WidgetEnabled               bool          `json:"widget_enabled,omitempty"`
	WidgetChannelID             string        `json:"widget_channel_id,omitempty"`
	VerificationLevel           int           `json:"verification_level,omitempty"`
	DefaultMessageNotifications int           `json:"default_message_notifications,omitempty"`
	ExplicitContentFilter       int           `json:"explicit_content_filter,omitempty"`
	Roles                       []*Role       `json:"roles"`
	Emojis                      []*Emoji      `json:"emojis"`
	Features                    []string      `json:"features"`
	MFALevel                    int           `json:"mfa_level,omitempty"`
	ApplicationID               string        `json:"application_id,omitempty"`
	SystemChannelID             string        `json:"system_channel_id,omitempty"`
	SystemChannelFlags          int           `json:"system_channel_flags,omitempty"`
	RulesChannelID              string        `json:"rules_channel_id,omitempty"`
	MaxPresences                int           `json:"max_presences,omitempty"`
	MaxMembers                  int           `json:"max_members,omitempty"`
	VanityURLCode               string        `json:"vanity_url_code,omitempty"`
	Banner                      string        `json:"banner,omitempty"`
	PremiumTier                 int           `json:"premium_tier,omitempty"`
	PremiumSubscriptionCount    int           `json:"premium_subscription_count,omitempty"`
	PreferredLocale             string        `json:"preferred_locale,omitempty"`
	PublicUpdatesChannelID      string        `json:"public_updates_channel_id,omitempty"`
	MaxVideoChannelUsers        int           `json:"max_video_channel_users,omitempty"`
	ApproximateMemberCount      int           `json:"approximate_member_count,omitempty"`
	ApproximatePresenceCount    int           `json:"approximate_presence_count,omitempty"`
	Channels                    []*Channel    `json:"channels,omitempty"`
	Members                     []*Member     `json:"members,omitempty"`
	MemberCount                 int           `json:"member_count,omitempty"`
	Large                       bool          `json:"large,omitempty"`
	Unavailable                 bool          `json:"unavailable,omitempty"`
}

// Channel represents a text or voice channel.
type Channel struct {
	ID                   string                 `json:"id"`
	Type                 int                    `json:"type"`
	GuildID              string                 `json:"guild_id,omitempty"`
	Position             int                    `json:"position,omitempty"`
	PermissionOverwrites []*PermissionOverwrite `json:"permission_overwrites,omitempty"`
	Name                 string                 `json:"name,omitempty"`
	Topic                string                 `json:"topic,omitempty"`
	NSFW                 bool                   `json:"nsfw,omitempty"`
	LastMessageID        string                 `json:"last_message_id,omitempty"`
	Bitrate              int                    `json:"bitrate,omitempty"`
	UserLimit            int                    `json:"user_limit,omitempty"`
	RateLimitPerUser     int                    `json:"rate_limit_per_user,omitempty"`
	Recipients           []*User                `json:"recipients,omitempty"`
	Icon                 string                 `json:"icon,omitempty"`
	OwnerID              string                 `json:"owner_id,omitempty"`
	ApplicationID        string                 `json:"application_id,omitempty"`
	ParentID             string                 `json:"parent_id,omitempty"`
	LastPinTimestamp     *time.Time             `json:"last_pin_timestamp,omitempty"`
}

// Mention returns the channel mention string.
func (c *Channel) Mention() string {
	return "<#" + c.ID + ">"
}

// PermissionOverwrite defines specific permissions for a role or user on a channel.
type PermissionOverwrite struct {
	ID    string `json:"id"`
	Type  int    `json:"type"` // 0 for role, 1 for member
	Allow string `json:"allow"`
	Deny  string `json:"deny"`
}

// Role represents a guild role.
type Role struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Color        int    `json:"color"`
	Hoist        bool   `json:"hoist"`
	Icon         string `json:"icon,omitempty"`
	UnicodeEmoji string `json:"unicode_emoji,omitempty"`
	Position     int    `json:"position"`
	Permissions  string `json:"permissions"`
	Managed      bool   `json:"managed"`
	Mentionable  bool   `json:"mentionable"`
	Flags        int    `json:"flags,omitempty"`
}

// Member represents a user in a specific guild.
type Member struct {
	GuildID      string    `json:"guild_id,omitempty"`
	User         *User     `json:"user"`
	Nick         string    `json:"nick,omitempty"`
	Avatar       string    `json:"avatar,omitempty"`
	Roles        []string  `json:"roles"`
	JoinedAt     time.Time `json:"joined_at"`
	PremiumSince time.Time `json:"premium_since,omitempty"`
	Deaf         bool      `json:"deaf"`
	Mute         bool      `json:"mute"`
	Pending      bool      `json:"pending,omitempty"`
	Permissions  string    `json:"permissions,omitempty"`
}

// Message represents a text message.
type Message struct {
	ID              string           `json:"id"`
	ChannelID       string           `json:"channel_id"`
	GuildID         string           `json:"guild_id,omitempty"`
	Author          *User            `json:"author"`
	Member          *Member          `json:"member,omitempty"`
	Content         string           `json:"content"`
	Timestamp       time.Time        `json:"timestamp"`
	EditedTimestamp *time.Time       `json:"edited_timestamp,omitempty"`
	TTS             bool             `json:"tts"`
	MentionEveryone bool             `json:"mention_everyone"`
	Mentions        []*User          `json:"mentions"`
	MentionRoles    []string         `json:"mention_roles"`
	Attachments     []*Attachment    `json:"attachments"`
	Embeds          []*Embed         `json:"embeds"`
	Reactions       []*Reaction      `json:"reactions,omitempty"`
	Nonce           interface{}      `json:"nonce,omitempty"`
	Pinned          bool             `json:"pinned"`
	WebhookID       string           `json:"webhook_id,omitempty"`
	Type            int              `json:"type"`
	MessageReference *MessageReference `json:"message_reference,omitempty"`
}

// MessageReference indicates reply context.
type MessageReference struct {
	MessageID string `json:"message_id,omitempty"`
	ChannelID string `json:"channel_id,omitempty"`
	GuildID   string `json:"guild_id,omitempty"`
}

// MessageSend payload for creating or editing messages.
type MessageSend struct {
	Content          string            `json:"content,omitempty"`
	TTS              bool              `json:"tts,omitempty"`
	Embeds           []*Embed          `json:"embeds,omitempty"`
	MessageReference *MessageReference `json:"message_reference,omitempty"`
}

// Embed structure for rich embed cards.
type Embed struct {
	Title       string          `json:"title,omitempty"`
	Type        string          `json:"type,omitempty"`
	Description string          `json:"description,omitempty"`
	URL         string          `json:"url,omitempty"`
	Timestamp   string          `json:"timestamp,omitempty"`
	Color       int             `json:"color,omitempty"`
	Footer      *EmbedFooter    `json:"footer,omitempty"`
	Image       *EmbedImage     `json:"image,omitempty"`
	Thumbnail   *EmbedThumbnail `json:"thumbnail,omitempty"`
	Video       *EmbedVideo     `json:"video,omitempty"`
	Provider    *EmbedProvider  `json:"provider,omitempty"`
	Author      *EmbedAuthor    `json:"author,omitempty"`
	Fields      []*EmbedField   `json:"fields,omitempty"`
}

// EmbedFooter structure.
type EmbedFooter struct {
	Text         string `json:"text"`
	IconURL      string `json:"icon_url,omitempty"`
	ProxyIconURL string `json:"proxy_icon_url,omitempty"`
}

// EmbedImage structure.
type EmbedImage struct {
	URL      string `json:"url"`
	ProxyURL string `json:"proxy_url,omitempty"`
	Height   int    `json:"height,omitempty"`
	Width    int    `json:"width,omitempty"`
}

// EmbedThumbnail structure.
type EmbedThumbnail struct {
	URL      string `json:"url"`
	ProxyURL string `json:"proxy_url,omitempty"`
	Height   int    `json:"height,omitempty"`
	Width    int    `json:"width,omitempty"`
}

// EmbedVideo structure.
type EmbedVideo struct {
	URL      string `json:"url,omitempty"`
	ProxyURL string `json:"proxy_url,omitempty"`
	Height   int    `json:"height,omitempty"`
	Width    int    `json:"width,omitempty"`
}

// EmbedProvider structure.
type EmbedProvider struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

// EmbedAuthor structure.
type EmbedAuthor struct {
	Name         string `json:"name,omitempty"`
	URL          string `json:"url,omitempty"`
	IconURL      string `json:"icon_url,omitempty"`
	ProxyIconURL string `json:"proxy_icon_url,omitempty"`
}

// EmbedField structure.
type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// Attachment file uploaded with message.
type Attachment struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	Description string `json:"description,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	Size        int    `json:"size"`
	URL         string `json:"url"`
	ProxyURL    string `json:"proxy_url"`
	Height      int    `json:"height,omitempty"`
	Width       int    `json:"width,omitempty"`
}

// Reaction represents a reaction to a message.
type Reaction struct {
	Count int    `json:"count"`
	Me    bool   `json:"me"`
	Emoji *Emoji `json:"emoji"`
}

// Emoji structure.
type Emoji struct {
	ID            string   `json:"id,omitempty"`
	Name          string   `json:"name"`
	Roles         []string `json:"roles,omitempty"`
	User          *User    `json:"user,omitempty"`
	RequireColons bool     `json:"require_colons,omitempty"`
	Managed       bool     `json:"managed,omitempty"`
	Animated      bool     `json:"animated,omitempty"`
	Available     bool     `json:"available,omitempty"`
}

// Invite structure.
type Invite struct {
	Code      string   `json:"code"`
	Guild     *Guild   `json:"guild,omitempty"`
	Channel   *Channel `json:"channel,omitempty"`
	Inviter   *User    `json:"inviter,omitempty"`
	Uses      int      `json:"uses,omitempty"`
	MaxUses   int      `json:"max_uses,omitempty"`
	MaxAge    int      `json:"max_age,omitempty"`
	Temporary bool     `json:"temporary,omitempty"`
	CreatedAt string   `json:"created_at,omitempty"`
}

// Webhook structure.
type Webhook struct {
	ID        string `json:"id"`
	Type      int    `json:"type"`
	GuildID   string `json:"guild_id,omitempty"`
	ChannelID string `json:"channel_id"`
	User      *User  `json:"user,omitempty"`
	Name      string `json:"name,omitempty"`
	Avatar    string `json:"avatar,omitempty"`
	Token     string `json:"token,omitempty"`
	URL       string `json:"url,omitempty"`
}

// GatewayPayload matches the wire packet schema: { op, d, s, t }.
type GatewayPayload struct {
	Op int             `json:"op"`
	D  json.RawMessage `json:"d,omitempty"`
	S  *int64          `json:"s,omitempty"`
	T  string          `json:"t,omitempty"`
}

// GatewayIdentify payload sent in OpIdentify.
type GatewayIdentify struct {
	Token          string                      `json:"token"`
	Properties     GatewayIdentifyProperties  `json:"properties"`
	Presence       *GatewayPresence           `json:"presence,omitempty"`
	Flags          int                         `json:"flags,omitempty"`
	InitialGuildID string                      `json:"initial_guild_id,omitempty"`
}

// GatewayIdentifyProperties standard client metadata.
type GatewayIdentifyProperties struct {
	OS      string `json:"$os"`
	Browser string `json:"$browser"`
	Device  string `json:"$device"`
}

// GatewayResume payload sent in OpResume.
type GatewayResume struct {
	Token     string `json:"token"`
	SessionID string `json:"session_id"`
	Seq       int64  `json:"seq"`
}

// GatewayHello payload received from OpHello.
type GatewayHello struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

// GatewayReady payload received on READY event.
type GatewayReady struct {
	V         int      `json:"v"`
	User      *User    `json:"user"`
	Guilds    []*Guild `json:"guilds"`
	SessionID string   `json:"session_id"`
	Shard     []int    `json:"shard,omitempty"`
}

// GatewayPresence status payload.
type GatewayPresence struct {
	Since      *int64              `json:"since"`
	Activities []*GatewayActivity `json:"activities"`
	Status     string              `json:"status"` // "online", "dnd", "idle", "invisible", "offline"
	AFK        bool                `json:"afk"`
}

// GatewayActivity activity payload.
type GatewayActivity struct {
	Name  string `json:"name"`
	Type  int    `json:"type"` // 0 = Game, 1 = Streaming, 2 = Listening, 3 = Watching, 4 = Custom, 5 = Competing
	URL   string `json:"url,omitempty"`
	State string `json:"state,omitempty"`
}

// GatewayError error payload received in OpGatewayError.
type GatewayError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
