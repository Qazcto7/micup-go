package micup

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

// VoiceStateUpdatePayload is sent in OpVoiceStateUpdate (Op 4).
type VoiceStateUpdatePayload struct {
	GuildID    string  `json:"guild_id"`
	ChannelID  *string `json:"channel_id"`
	SelfMute   bool    `json:"self_mute"`
	SelfDeaf   bool    `json:"self_deaf"`
	SelfVideo  bool    `json:"self_video,omitempty"`
	SelfStream bool    `json:"self_stream,omitempty"`
}

// VoiceConnection represents an active voice session in a guild channel.
type VoiceConnection struct {
	GuildID   string
	ChannelID string
	UserID    string
	SessionID string
	Token     string
	Endpoint  string
	Ready     chan struct{}
	mu        sync.RWMutex
	session   *Session
}

// VoiceManager tracks active voice connections per guild.
type VoiceManager struct {
	connections map[string]*VoiceConnection
	mu          sync.RWMutex
}

// NewVoiceManager initializes a voice manager.
func NewVoiceManager() *VoiceManager {
	return &VoiceManager{
		connections: make(map[string]*VoiceConnection),
	}
}

// Get retrieves an existing voice connection for a guild.
func (v *VoiceManager) Get(guildID string) *VoiceConnection {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.connections[guildID]
}

// Set stores a voice connection.
func (v *VoiceManager) Set(guildID string, vc *VoiceConnection) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.connections[guildID] = vc
}

// Remove deletes a voice connection.
func (v *VoiceManager) Remove(guildID string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	delete(v.connections, guildID)
}

// ChannelVoiceJoin instructs the gateway to join a voice channel in a guild.
func (s *Session) ChannelVoiceJoin(ctx context.Context, guildID, channelID string, mute, deaf bool) (*VoiceConnection, error) {
	vc := &VoiceConnection{
		GuildID:   guildID,
		ChannelID: channelID,
		Ready:     make(chan struct{}),
		session:   s,
	}

	payload := VoiceStateUpdatePayload{
		GuildID:   guildID,
		ChannelID: &channelID,
		SelfMute:  mute,
		SelfDeaf:  deaf,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	err = s.gateway.SendPayload(&GatewayPayload{
		Op: OpVoiceStateUpdate,
		D:  data,
	})
	if err != nil {
		return nil, err
	}

	return vc, nil
}

// ChannelVoiceLeave leaves the voice channel in a guild.
func (s *Session) ChannelVoiceLeave(guildID string) error {
	payload := VoiceStateUpdatePayload{
		GuildID:   guildID,
		ChannelID: nil, // null instructs server to disconnect voice
		SelfMute:  false,
		SelfDeaf:  false,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return s.gateway.SendPayload(&GatewayPayload{
		Op: OpVoiceStateUpdate,
		D:  data,
	})
}

// Disconnect disconnects the voice connection.
func (vc *VoiceConnection) Disconnect() error {
	return vc.session.ChannelVoiceLeave(vc.GuildID)
}

// WaitUntilReady blocks until both VoiceState and VoiceServer tokens have arrived or context expires.
func (vc *VoiceConnection) WaitUntilReady(ctx context.Context) error {
	select {
	case <-vc.Ready:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(15 * time.Second):
		return errors.New("timeout waiting for voice connection credentials")
	}
}
