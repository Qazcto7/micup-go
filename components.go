package micup

import "encoding/json"

// ComponentType constants.
const (
	ComponentTypeActionRow  = 1
	ComponentTypeButton     = 2
	ComponentTypeSelectMenu = 3
	ComponentTypeTextInput  = 4
)

// ButtonStyle constants.
const (
	ButtonStylePrimary   = 1 // Brand/Accent color
	ButtonStyleSecondary = 2 // Grey / Neutral
	ButtonStyleSuccess   = 3 // Green
	ButtonStyleDanger    = 4 // Red
	ButtonStyleLink      = 5 // Link to external URL
)

// Component represents an interactive UI component.
type Component interface {
	Type() int
}

// ActionRow holds a row of child components (like buttons).
type ActionRow struct {
	Components []Component `json:"components"`
}

func (a *ActionRow) Type() int {
	return ComponentTypeActionRow
}

// MarshalJSON custom serializer for ActionRow.
func (a *ActionRow) MarshalJSON() ([]byte, error) {
	type Alias ActionRow
	return json.Marshal(struct {
		Type int `json:"type"`
		*Alias
	}{
		Type:  ComponentTypeActionRow,
		Alias: (*Alias)(a),
	})
}

// Button represents an interactive clickable button.
type Button struct {
	Style    int    `json:"style"`
	Label    string `json:"label,omitempty"`
	Emoji    *Emoji `json:"emoji,omitempty"`
	CustomID string `json:"custom_id,omitempty"`
	URL      string `json:"url,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
}

func (b *Button) Type() int {
	return ComponentTypeButton
}

// MarshalJSON custom serializer for Button.
func (b *Button) MarshalJSON() ([]byte, error) {
	type Alias Button
	return json.Marshal(struct {
		Type int `json:"type"`
		*Alias
	}{
		Type:  ComponentTypeButton,
		Alias: (*Alias)(b),
	})
}

// Helper constructors

// WithDisabled sets the disabled state of the button.
func (b *Button) WithDisabled(disabled bool) *Button {
	b.Disabled = disabled
	return b
}

// WithEmoji sets a unicode or custom emoji on the button.
func (b *Button) WithEmoji(emojiName string) *Button {
	b.Emoji = &Emoji{Name: emojiName}
	return b
}

// WithLabel sets or updates the label of the button.
func (b *Button) WithLabel(label string) *Button {
	b.Label = label
	return b
}

// WithURL sets the external URL for a link button.
func (b *Button) WithURL(url string) *Button {
	b.URL = url
	return b
}

// NewButton creates a button with custom style, customID and label.
func NewButton(style int, customID, label string) *Button {
	return &Button{
		Style:    style,
		CustomID: customID,
		Label:    label,
	}
}

// NewPrimaryButton creates a primary styled button.
func NewPrimaryButton(customID, label string) *Button {
	return &Button{
		Style:    ButtonStylePrimary,
		CustomID: customID,
		Label:    label,
	}
}

// NewSecondaryButton creates a secondary neutral button.
func NewSecondaryButton(customID, label string) *Button {
	return &Button{
		Style:    ButtonStyleSecondary,
		CustomID: customID,
		Label:    label,
	}
}

// NewSuccessButton creates a green success button.
func NewSuccessButton(customID, label string) *Button {
	return &Button{
		Style:    ButtonStyleSuccess,
		CustomID: customID,
		Label:    label,
	}
}

// NewDangerButton creates a red danger button.
func NewDangerButton(customID, label string) *Button {
	return &Button{
		Style:    ButtonStyleDanger,
		CustomID: customID,
		Label:    label,
	}
}

// NewLinkButton creates a button that opens an external link.
func NewLinkButton(url, label string) *Button {
	return &Button{
		Style: ButtonStyleLink,
		URL:   url,
		Label: label,
	}
}

// NewActionRow wraps buttons in an ActionRow.
func NewActionRow(components ...Component) *ActionRow {
	return &ActionRow{
		Components: components,
	}
}
