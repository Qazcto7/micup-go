package micup

import (
	"context"
	"fmt"
)

// ApplicationCommandType defines the type of command.
const (
	ChatInputCommandType = 1
	UserCommandType       = 2
	MessageCommandType    = 3
)

// ApplicationCommandOptionType defines argument types for commands.
const (
	OptionTypeSubCommand      = 1
	OptionTypeSubCommandGroup = 2
	OptionTypeString          = 3
	OptionTypeInteger         = 4
	OptionTypeBoolean         = 5
	OptionTypeUser            = 6
	OptionTypeChannel         = 7
	OptionTypeRole            = 8
	OptionTypeMentionable     = 9
	OptionTypeNumber          = 10
	OptionTypeAttachment      = 11
)

// InteractionType defines incoming interaction kinds.
const (
	InteractionPing                           = 1
	InteractionApplicationCommand             = 2
	InteractionMessageComponent               = 3
	InteractionApplicationCommandAutocomplete = 4
	InteractionModalSubmit                    = 5
)

// InteractionResponseType defines responses to interactions.
const (
	InteractionResponsePong                                 = 1
	InteractionResponseChannelMessageWithSource             = 4
	InteractionResponseDeferredChannelMessageWithSource     = 5
	InteractionResponseDeferredUpdateMessage                 = 6
	InteractionResponseUpdateMessage                         = 7
	InteractionResponseApplicationCommandAutocompleteResult = 8
	InteractionResponseModal                                = 9
)

// ApplicationCommand represents a registered slash command.
type ApplicationCommand struct {
	ID                       string                      `json:"id,omitempty"`
	ApplicationID            string                      `json:"application_id,omitempty"`
	Type                     int                         `json:"type,omitempty"`
	Name                     string                      `json:"name"`
	Description              string                      `json:"description"`
	Options                  []*ApplicationCommandOption `json:"options,omitempty"`
	DefaultMemberPermissions *string                     `json:"default_member_permissions,omitempty"`
	DMPermission             *bool                       `json:"dm_permission,omitempty"`
}

// ApplicationCommandOption represents a parameter for a slash command.
type ApplicationCommandOption struct {
	Type        int                               `json:"type"`
	Name        string                            `json:"name"`
	Description string                            `json:"description"`
	Required    bool                              `json:"required,omitempty"`
	Choices     []*ApplicationCommandOptionChoice `json:"choices,omitempty"`
	Options     []*ApplicationCommandOption       `json:"options,omitempty"`
	MinValue    *float64                          `json:"min_value,omitempty"`
	MaxValue    *float64                          `json:"max_value,omitempty"`
}

// ApplicationCommandOptionChoice defines predefined choice options.
type ApplicationCommandOptionChoice struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"` // string, int, or float
}

// Interaction represents an incoming interaction payload (Slash command invoke, button click, etc.).
type Interaction struct {
	ID            string           `json:"id"`
	ApplicationID string           `json:"application_id"`
	Type          int              `json:"type"`
	Data          *InteractionData `json:"data,omitempty"`
	GuildID       string           `json:"guild_id,omitempty"`
	ChannelID     string           `json:"channel_id,omitempty"`
	Member        *Member          `json:"member,omitempty"`
	User          *User            `json:"user,omitempty"`
	Token         string           `json:"token"`
	Version       int              `json:"version"`
	Message       *Message         `json:"message,omitempty"`
}

// InteractionData details about the command or component invoked.
type InteractionData struct {
	ID            string                   `json:"id"`
	Name          string                   `json:"name"`
	Type          int                      `json:"type"`
	Options       []*InteractionDataOption `json:"options,omitempty"`
	CustomID      string                   `json:"custom_id,omitempty"`
	ComponentType int                      `json:"component_type,omitempty"`
	Values        []string                 `json:"values,omitempty"`
}

// InteractionDataOption represents an argument passed to a command.
type InteractionDataOption struct {
	Name    string                   `json:"name"`
	Type    int                      `json:"type"`
	Value   interface{}              `json:"value,omitempty"`
	Options []*InteractionDataOption `json:"options,omitempty"`
	Focused bool                     `json:"focused,omitempty"`
}

// InteractionResponse represents the payload sent back to acknowledge an interaction.
type InteractionResponse struct {
	Type int                      `json:"type"`
	Data *InteractionResponseData `json:"data,omitempty"`
}

// InteractionResponseData contains response message content or modal fields.
type InteractionResponseData struct {
	TTS        bool         `json:"tts,omitempty"`
	Content    string       `json:"content,omitempty"`
	Embeds     []*Embed     `json:"embeds,omitempty"`
	Flags      int          `json:"flags,omitempty"` // 64 = Ephemeral (visible only to user)
	Components []Component  `json:"components,omitempty"`
}

// REST methods for Application Commands

// CreateGlobalApplicationCommand registers a global application slash command.
func (c *RESTClient) CreateGlobalApplicationCommand(ctx context.Context, appID string, cmd *ApplicationCommand) (*ApplicationCommand, error) {
	var created ApplicationCommand
	err := c.RequestJSON(ctx, "POST", fmt.Sprintf("/applications/%s/commands", appID), cmd, &created)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

// GetGlobalApplicationCommands fetches all registered global commands.
func (c *RESTClient) GetGlobalApplicationCommands(ctx context.Context, appID string) ([]*ApplicationCommand, error) {
	var cmds []*ApplicationCommand
	err := c.RequestJSON(ctx, "GET", fmt.Sprintf("/applications/%s/commands", appID), nil, &cmds)
	if err != nil {
		return nil, err
	}
	return cmds, nil
}

// DeleteGlobalApplicationCommand deletes a global command.
func (c *RESTClient) DeleteGlobalApplicationCommand(ctx context.Context, appID, cmdID string) error {
	return c.RequestJSON(ctx, "DELETE", fmt.Sprintf("/applications/%s/commands/%s", appID, cmdID), nil, nil)
}

// CreateGuildApplicationCommand registers a slash command for a specific guild.
func (c *RESTClient) CreateGuildApplicationCommand(ctx context.Context, appID, guildID string, cmd *ApplicationCommand) (*ApplicationCommand, error) {
	var created ApplicationCommand
	err := c.RequestJSON(ctx, "POST", fmt.Sprintf("/applications/%s/guilds/%s/commands", appID, guildID), cmd, &created)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

// GetGuildApplicationCommands fetches commands registered for a guild.
func (c *RESTClient) GetGuildApplicationCommands(ctx context.Context, appID, guildID string) ([]*ApplicationCommand, error) {
	var cmds []*ApplicationCommand
	err := c.RequestJSON(ctx, "GET", fmt.Sprintf("/applications/%s/guilds/%s/commands", appID, guildID), nil, &cmds)
	if err != nil {
		return nil, err
	}
	return cmds, nil
}

// DeleteGuildApplicationCommand deletes a guild-scoped command.
func (c *RESTClient) DeleteGuildApplicationCommand(ctx context.Context, appID, guildID, cmdID string) error {
	return c.RequestJSON(ctx, "DELETE", fmt.Sprintf("/applications/%s/guilds/%s/commands/%s", appID, guildID, cmdID), nil, nil)
}

// InteractionRespond sends a response back to an interaction.
func (c *RESTClient) InteractionRespond(ctx context.Context, interactionID, token string, resp *InteractionResponse) error {
	endpoint := fmt.Sprintf("/interactions/%s/%s/callback", interactionID, token)
	return c.RequestJSON(ctx, "POST", endpoint, resp, nil)
}

// Session convenience methods

// CreateGlobalApplicationCommand registers a global command.
func (s *Session) CreateGlobalApplicationCommand(appID string, cmd *ApplicationCommand) (*ApplicationCommand, error) {
	return s.rest.CreateGlobalApplicationCommand(context.Background(), appID, cmd)
}

// CreateGuildApplicationCommand registers a guild command.
func (s *Session) CreateGuildApplicationCommand(appID, guildID string, cmd *ApplicationCommand) (*ApplicationCommand, error) {
	return s.rest.CreateGuildApplicationCommand(context.Background(), appID, guildID, cmd)
}

// InteractionRespond responds to an interaction.
func (s *Session) InteractionRespond(interaction *Interaction, resp *InteractionResponse) error {
	return s.rest.InteractionRespond(context.Background(), interaction.ID, interaction.Token, resp)
}

// RespondText responds to an interaction with a simple text message.
func (s *Session) RespondText(interaction *Interaction, text string) error {
	return s.InteractionRespond(interaction, &InteractionResponse{
		Type: InteractionResponseChannelMessageWithSource,
		Data: &InteractionResponseData{
			Content: text,
		},
	})
}
