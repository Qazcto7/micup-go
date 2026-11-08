package micup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// VoiceTrackKind constants.
const (
	TrackKindAudio = "audio"
	TrackKindVideo = "video"
)

// VoiceTrackSource constants matching @fluxer/voice_engine_v2 reverse engineering.
const (
	TrackSourceMicrophone  = "microphone"
	TrackSourceScreen      = "screen"
	TrackSourceScreenAudio = "screenAudio"
	TrackSourceCamera      = "camera"
)

// VoiceCodec constants.
const (
	CodecOpus = "opus"
	CodecH264 = "h264"
	CodecVP8  = "vp8"
	CodecVP9  = "vp9"
	CodecAV1  = "av1"
)

// VoiceTrack represents an active audio or video media track in a room.
type VoiceTrack struct {
	SID                 string `json:"sid"`
	ParticipantIdentity string `json:"participant_identity"`
	ParticipantSID      string `json:"participant_sid,omitempty"`
	Kind                string `json:"kind"`                // "audio", "video"
	Source              string `json:"source"`              // "microphone", "screen", "screenAudio", "camera"
	Codec               string `json:"codec"`               // "opus", "h264", "vp8", "vp9"
	BitrateKbps         int    `json:"bitrate_kbps"`        // 48, 128, etc.
	SampleRate          int    `json:"sample_rate"`         // 48000
	Channels            int    `json:"channels"`            // 1 (mono), 2 (stereo)
	Muted               bool   `json:"muted"`
}

// VoiceParticipant represents a member in the voice channel room.
type VoiceParticipant struct {
	SID        string                 `json:"sid"`
	Identity   string                 `json:"identity"` // UserID
	Name       string                 `json:"name"`
	Speaking   bool                   `json:"speaking"`
	AudioLevel float64                `json:"audio_level"`
	Tracks     map[string]*VoiceTrack `json:"tracks"`
	JoinedAt   time.Time              `json:"joined_at"`
}

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
	mu           sync.RWMutex
	session      *Session
	GuildID      string
	ChannelID    string
	RoomName     string
	UserID       string
	SessionID    string
	Token        string
	Endpoint     string
	Ready        chan struct{}
	Streamer     *AudioStreamer
	Participants map[string]*VoiceParticipant // userID -> VoiceParticipant

	// Callbacks
	onSpeaking          func(userID string, speaking bool)
	onParticipantJoined func(p *VoiceParticipant)
	onParticipantLeft   func(userID string)
	onTrackPublished    func(t *VoiceTrack)
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

// VoiceJoin is a convenience helper to join a voice channel with default context.
func (s *Session) VoiceJoin(guildID, channelID string, mute, deaf bool) (*VoiceConnection, error) {
	return s.ChannelVoiceJoin(context.Background(), guildID, channelID, mute, deaf)
}

// ChannelVoiceJoin instructs the gateway to join a voice channel in a guild.
func (s *Session) ChannelVoiceJoin(ctx context.Context, guildID, channelID string, mute, deaf bool) (*VoiceConnection, error) {
	vc := &VoiceConnection{
		GuildID:      guildID,
		ChannelID:    channelID,
		RoomName:     fmt.Sprintf("%s/%s", guildID, channelID),
		Ready:        make(chan struct{}),
		session:      s,
		Participants: make(map[string]*VoiceParticipant),
		Streamer:     NewAudioStreamer(nil),
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

// Disconnect disconnects the voice connection and closes media streams.
func (vc *VoiceConnection) Disconnect() error {
	vc.mu.Lock()
	if vc.Streamer != nil {
		vc.Streamer.Close()
	}
	vc.mu.Unlock()

	return vc.session.ChannelVoiceLeave(vc.GuildID)
}

// SendOpusFrame transmits an Opus audio packet to the voice room.
func (vc *VoiceConnection) SendOpusFrame(frame []byte) error {
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	if vc.Streamer == nil {
		return errors.New("audio streamer not initialized")
	}
	return vc.Streamer.SendOpusFrame(frame)
}

// PublishAudioTrack registers and prepares an audio track for streaming.
func (vc *VoiceConnection) PublishAudioTrack(source string, sampleRate, channels, bitrateKbps int) (*VoiceTrack, error) {
	if sampleRate <= 0 {
		sampleRate = 48000
	}
	if channels <= 0 {
		channels = 2
	}
	if bitrateKbps <= 0 {
		bitrateKbps = 48
	}

	track := &VoiceTrack{
		SID:                 fmt.Sprintf("track-audio-%d", time.Now().UnixNano()),
		ParticipantIdentity: vc.UserID,
		Kind:                TrackKindAudio,
		Source:              source,
		Codec:               CodecOpus,
		BitrateKbps:         bitrateKbps,
		SampleRate:          sampleRate,
		Channels:            channels,
	}

	vc.mu.Lock()
	if vc.onTrackPublished != nil {
		vc.onTrackPublished(track)
	}
	vc.mu.Unlock()

	return track, nil
}

// SetSpeaking updates the local bot speaking indicator.
func (vc *VoiceConnection) SetSpeaking(speaking bool) error {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	if vc.onSpeaking != nil {
		vc.onSpeaking(vc.UserID, speaking)
	}
	return nil
}

// Event handler setters

// OnSpeaking sets callback for participant speaking events.
func (vc *VoiceConnection) OnSpeaking(handler func(userID string, speaking bool)) {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	vc.onSpeaking = handler
}

// OnParticipantJoined sets callback when a user joins the voice room.
func (vc *VoiceConnection) OnParticipantJoined(handler func(p *VoiceParticipant)) {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	vc.onParticipantJoined = handler
}

// OnParticipantLeft sets callback when a user leaves the voice room.
func (vc *VoiceConnection) OnParticipantLeft(handler func(userID string)) {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	vc.onParticipantLeft = handler
}

// OnTrackPublished sets callback when an audio/video track publishes.
func (vc *VoiceConnection) OnTrackPublished(handler func(t *VoiceTrack)) {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	vc.onTrackPublished = handler
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
