package micup

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// TicketConfig defines settings for the ticket support system.
type TicketConfig struct {
	GuildID            string   `json:"guild_id"`
	CategoryID         string   `json:"category_id,omitempty"`
	SupportRoleIDs     []string `json:"support_role_ids"`
	ChannelPrefix      string   `json:"channel_prefix"`       // e.g. "destek-" or "ticket-"
	OpenButtonID       string   `json:"open_button_id"`       // CustomID for opening ticket
	CloseButtonID      string   `json:"close_button_id"`      // CustomID for closing ticket
	LogChannelID       string   `json:"log_channel_id,omitempty"` // Channel to send ticket close logs
	MaxTicketsPerUser  int      `json:"max_tickets_per_user"`
}

// Ticket represents an active support ticket.
type Ticket struct {
	ID        string    `json:"id"`
	ChannelID string    `json:"channel_id"`
	GuildID   string    `json:"guild_id"`
	CreatorID string    `json:"creator_id"`
	CreatedAt time.Time `json:"created_at"`
	Closed    bool      `json:"closed"`
}

// TicketManager manages ticket creation, lifecycle, and permissions.
type TicketManager struct {
	config  *TicketConfig
	session *Session
	tickets map[string]*Ticket // channelID -> Ticket
	mu      sync.RWMutex
}

// NewTicketManager initializes a ticket system with the given configuration.
func NewTicketManager(session *Session, cfg *TicketConfig) *TicketManager {
	if cfg.ChannelPrefix == "" {
		cfg.ChannelPrefix = "ticket-"
	}
	if cfg.OpenButtonID == "" {
		cfg.OpenButtonID = "ticket_open"
	}
	if cfg.CloseButtonID == "" {
		cfg.CloseButtonID = "ticket_close"
	}
	if cfg.MaxTicketsPerUser <= 0 {
		cfg.MaxTicketsPerUser = 1
	}

	return &TicketManager{
		config:  cfg,
		session: session,
		tickets: make(map[string]*Ticket),
	}
}

// CreatePanelMessage creates an embed and button row for users to open a ticket.
func (tm *TicketManager) CreatePanelMessage(title, description, buttonLabel string) (*Embed, *ActionRow) {
	if title == "" {
		title = "🎫 Destek Talebi Oluştur"
	}
	if description == "" {
		description = "Yetkili ekibimizle özel olarak görüşmek ve destek almak için aşağıdaki butona tıklayın."
	}
	if buttonLabel == "" {
		buttonLabel = "Destek Talebi Aç"
	}

	embed := NewEmbed().
		SetTitle(title).
		SetDescription(description).
		SetColor(0x3498db).
		SetFooter("Destek Sistemi", "")

	btn := NewPrimaryButton(tm.config.OpenButtonID, buttonLabel).WithEmoji("📩")
	row := NewActionRow(btn)

	return embed.Build(), row
}

// HandleInteraction processes button clicks related to opening or closing tickets.
// Returns whether the interaction was handled.
func (tm *TicketManager) HandleInteraction(i *Interaction) (bool, error) {
	if i == nil || i.Data == nil {
		return false, nil
	}

	customID := i.Data.CustomID

	switch customID {
	case tm.config.OpenButtonID:
		return true, tm.handleOpenTicket(i)
	case tm.config.CloseButtonID:
		return true, tm.handleCloseTicket(i)
	default:
		return false, nil
	}
}

func (tm *TicketManager) handleOpenTicket(i *Interaction) error {
	user := i.User
	if user == nil && i.Member != nil {
		user = i.Member.User
	}
	if user == nil {
		return fmt.Errorf("user not found on interaction")
	}

	tm.mu.RLock()
	activeCount := 0
	for _, t := range tm.tickets {
		if t.CreatorID == user.ID && !t.Closed {
			activeCount++
		}
	}
	tm.mu.RUnlock()

	if activeCount >= tm.config.MaxTicketsPerUser {
		return tm.session.RespondText(i, "⚠️ Zaten açık bir destek talebiniz bulunuyor!")
	}

	// Channel permission overwrites:
	// 1. @everyone cannot see channel
	// 2. Creator can see and send messages
	// 3. Support roles can see and manage messages
	overwrites := []*PermissionOverwrite{
		{
			ID:    tm.config.GuildID, // @everyone
			Type:  0,                 // role
			Deny:  FormatPermissions(PermissionViewChannel),
			Allow: "0",
		},
		{
			ID:    user.ID,
			Type:  1, // member
			Allow: FormatPermissions(PermissionViewChannel | PermissionSendMessages | PermissionAttachFiles | PermissionReadMessageHistory),
			Deny:  "0",
		},
	}

	for _, roleID := range tm.config.SupportRoleIDs {
		overwrites = append(overwrites, &PermissionOverwrite{
			ID:    roleID,
			Type:  0, // role
			Allow: FormatPermissions(PermissionViewChannel | PermissionSendMessages | PermissionManageMessages | PermissionReadMessageHistory),
			Deny:  "0",
		})
	}

	channelName := fmt.Sprintf("%s%s", tm.config.ChannelPrefix, strings.ToLower(user.Username))
	newChannel, err := tm.session.GuildChannelCreate(tm.config.GuildID, &ChannelParams{
		Name:                 channelName,
		Type:                 ChannelTypeGuildText,
		ParentID:             tm.config.CategoryID,
		PermissionOverwrites: overwrites,
	})
	if err != nil {
		_ = tm.session.RespondText(i, fmt.Sprintf("Destek kanalı oluşturulamadı: %v", err))
		return err
	}

	ticket := &Ticket{
		ID:        fmt.Sprintf("t-%d", time.Now().UnixNano()),
		ChannelID: newChannel.ID,
		GuildID:   tm.config.GuildID,
		CreatorID: user.ID,
		CreatedAt: time.Now(),
	}

	tm.mu.Lock()
	tm.tickets[newChannel.ID] = ticket
	tm.mu.Unlock()

	// Send initial message in ticket channel
	closeBtn := NewDangerButton(tm.config.CloseButtonID, "Talebi Kapat").WithEmoji("🔒")
	welcomeEmbed := NewEmbed().
		SetTitle(fmt.Sprintf("🎫 Destek Talebi #%s", newChannel.Name)).
		SetDescription(fmt.Sprintf("Hoş geldiniz <@%s>!\nYetkili ekibimiz en kısa sürede size yardımcı olacaktır.\nLütfen sorununuzu detaylı bir şekilde açıklayın.\n\nİşleminiz bittiğinde aşağıdaki butondan talebi kapatabilirsiniz.", user.ID)).
		SetColor(0x2ecc71).
		AddField("Açan Kullanıcı", fmt.Sprintf("<@%s> (%s)", user.ID, user.Username), true).
		AddField("Açılış Zamanı", ticket.CreatedAt.Format("15:04:05 02.01.2006"), true).
		Build()

	_, _ = tm.session.ChannelMessageSendComplex(newChannel.ID, &MessageSend{
		Content:    fmt.Sprintf("<@%s> Talebiniz oluşturuldu.", user.ID),
		Embeds:     []*Embed{welcomeEmbed},
		Components: []Component{NewActionRow(closeBtn)},
	})

	// Ephemeral or direct response in interaction channel
	return tm.session.RespondText(i, fmt.Sprintf("✅ Destek talebiniz oluşturuldu: <#%s>", newChannel.ID))
}

func (tm *TicketManager) handleCloseTicket(i *Interaction) error {
	user := i.User
	if user == nil && i.Member != nil {
		user = i.Member.User
	}
	closerName := "Yetkili"
	if user != nil {
		closerName = user.Username
	}

	channelID := i.ChannelID

	tm.mu.Lock()
	ticket, exists := tm.tickets[channelID]
	if exists {
		ticket.Closed = true
	}
	tm.mu.Unlock()

	_ = tm.session.RespondText(i, "🔒 Destek talebi kapatılıyor ve kanal siliniyor...")

	// Log if channel configured
	if tm.config.LogChannelID != "" && ticket != nil {
		duration := time.Since(ticket.CreatedAt).Round(time.Second)
		logEmbed := NewEmbed().
			SetTitle("📋 Destek Talebi Kapatıldı").
			SetColor(0xe74c3c).
			AddField("Kanal", channelID, true).
			AddField("Açan Kullanıcı", fmt.Sprintf("<@%s>", ticket.CreatorID), true).
			AddField("Kapatan", closerName, true).
			AddField("Açık Kalma Süresi", duration.String(), true).
			Build()

		_, _ = tm.session.ChannelMessageSendEmbed(tm.config.LogChannelID, logEmbed)
	}

	// Delete channel after a short delay
	go func() {
		time.Sleep(3 * time.Second)
		_ = tm.session.ChannelDelete(channelID)
		tm.mu.Lock()
		delete(tm.tickets, channelID)
		tm.mu.Unlock()
	}()

	return nil
}

// ActiveTickets returns the number of open tickets.
func (tm *TicketManager) ActiveTickets() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	count := 0
	for _, t := range tm.tickets {
		if !t.Closed {
			count++
		}
	}
	return count
}
