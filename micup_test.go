package micup

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestOpCodes(t *testing.T) {
	if OpDispatch != 0 {
		t.Fatalf("expected OpDispatch to be 0, got %d", OpDispatch)
	}
	if OpHeartbeat != 1 {
		t.Fatalf("expected OpHeartbeat to be 1, got %d", OpHeartbeat)
	}
	if OpIdentify != 2 {
		t.Fatalf("expected OpIdentify to be 2, got %d", OpIdentify)
	}
	if OpHello != 10 {
		t.Fatalf("expected OpHello to be 10, got %d", OpHello)
	}
	if OpHeartbeatAck != 11 {
		t.Fatalf("expected OpHeartbeatAck to be 11, got %d", OpHeartbeatAck)
	}
	if OpGatewayError != 12 {
		t.Fatalf("expected OpGatewayError to be 12, got %d", OpGatewayError)
	}
}

func TestDiscovery(t *testing.T) {
	mockResponse := DiscoveryManifest{
		APICodeVersion: 1,
		Endpoints: DiscoveryEndpoints{
			API:     "https://test.micup.gg/api",
			Gateway: "wss://test.micup.gg/gateway",
			Media:   "https://media.test.micup.gg",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != DefaultWellKnown {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer ts.Close()

	disc := NewDiscoveryManager(ts.URL, ts.Client())
	manifest, err := disc.Discover(context.Background())
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}

	if manifest.APICodeVersion != 1 {
		t.Errorf("expected api version 1, got %d", manifest.APICodeVersion)
	}
	if manifest.Endpoints.API != "https://test.micup.gg/api" {
		t.Errorf("expected api endpoint https://test.micup.gg/api, got %s", manifest.Endpoints.API)
	}
	if manifest.Endpoints.Gateway != "wss://test.micup.gg/gateway" {
		t.Errorf("expected gateway endpoint wss://test.micup.gg/gateway, got %s", manifest.Endpoints.Gateway)
	}
}

func TestEventHandlerRegistry(t *testing.T) {
	registry := NewEventHandlerRegistry()

	var received atomic.Bool
	unregister := registry.Add(func(m *MessageCreate) {
		if m.Content == "ping" {
			received.Store(true)
		}
	})

	registry.Dispatch(nil, &MessageCreate{
		Message: &Message{Content: "ping"},
	})

	time.Sleep(50 * time.Millisecond)
	if !received.Load() {
		t.Error("expected event to be received by handler")
	}

	// Test unregister
	unregister()
	received.Store(false)
	registry.Dispatch(nil, &MessageCreate{
		Message: &Message{Content: "ping"},
	})
	time.Sleep(50 * time.Millisecond)
	if received.Load() {
		t.Error("expected handler to not be called after unregister")
	}
}

func TestCommandRouter(t *testing.T) {
	router := NewCommandRouter("!")

	executed := false
	var passedArg string

	router.Register(&Command{
		Name:        "test",
		Description: "A unit test command",
		Handler: func(ctx *CommandContext) error {
			executed = true
			if len(ctx.Args) > 0 {
				passedArg = ctx.Args[0]
			}
			return nil
		},
	})

	// Test message without prefix
	router.HandleMessage(nil, &Message{
		Content: "test hello",
		Author:  &User{Bot: false},
	})
	if executed {
		t.Error("command should not execute without prefix")
	}

	// Test message from bot
	router.HandleMessage(nil, &Message{
		Content: "!test hello",
		Author:  &User{Bot: true},
	})
	if executed {
		t.Error("command should ignore bot authors")
	}

	// Test valid command
	router.HandleMessage(nil, &Message{
		Content: "!test foobar",
		Author:  &User{Bot: false},
	})
	if !executed {
		t.Error("command should execute on valid message")
	}
	if passedArg != "foobar" {
		t.Errorf("expected passedArg 'foobar', got '%s'", passedArg)
	}
}

func TestPayloadSerialization(t *testing.T) {
	seq := int64(42)
	payload := GatewayPayload{
		Op: OpDispatch,
		S:  &seq,
		T:  EventMessageCreate,
		D:  json.RawMessage(`{"content":"test"}`),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json marshal failed: %v", err)
	}

	var parsed GatewayPayload
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}

	if parsed.Op != OpDispatch {
		t.Errorf("expected op %d, got %d", OpDispatch, parsed.Op)
	}
	if *parsed.S != 42 {
		t.Errorf("expected sequence 42, got %d", *parsed.S)
	}
	if parsed.T != EventMessageCreate {
		t.Errorf("expected type %s, got %s", EventMessageCreate, parsed.T)
	}
}

func TestEmbedBuilder(t *testing.T) {
	embed := NewEmbedBuilder().
		SetTitle("Test Title").
		SetDescription("Test Desc").
		SetColor(0xFF5500).
		AddField("Field 1", "Value 1", true).
		SetFooter("Footer", "http://icon.png").
		Build()

	if embed.Title != "Test Title" {
		t.Errorf("expected Title 'Test Title', got '%s'", embed.Title)
	}
	if embed.Color != 0xFF5500 {
		t.Errorf("expected Color 0xFF5500, got %d", embed.Color)
	}
	if len(embed.Fields) != 1 || embed.Fields[0].Name != "Field 1" {
		t.Errorf("expected 1 field named 'Field 1'")
	}
}

func TestComponentSerialization(t *testing.T) {
	btn := NewPrimaryButton("btn_click", "Click Me")
	row := NewActionRow(btn)

	data, err := json.Marshal(row)
	if err != nil {
		t.Fatalf("failed to marshal ActionRow: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal ActionRow json: %v", err)
	}

	if raw["type"] != float64(ComponentTypeActionRow) {
		t.Errorf("expected type 1, got %v", raw["type"])
	}
}

func TestVoiceStatePayload(t *testing.T) {
	chID := "ch123"
	payload := VoiceStateUpdatePayload{
		GuildID:   "g123",
		ChannelID: &chID,
		SelfMute:  true,
		SelfDeaf:  false,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal voice state payload: %v", err)
	}

	var parsed VoiceStateUpdatePayload
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal voice state payload: %v", err)
	}

	if *parsed.ChannelID != "ch123" || !parsed.SelfMute {
		t.Errorf("unexpected parsed voice payload: %+v", parsed)
	}
}

func TestVoiceTracksAndParticipants(t *testing.T) {
	vc := &VoiceConnection{
		GuildID:      "g1",
		ChannelID:    "c1",
		RoomName:     "g1/c1",
		UserID:       "u1",
		Streamer:     NewAudioStreamer(nil),
		Participants: make(map[string]*VoiceParticipant),
	}
	defer vc.Streamer.Close()

	if vc.RoomName != "g1/c1" {
		t.Errorf("expected room name g1/c1, got %s", vc.RoomName)
	}

	track, err := vc.PublishAudioTrack(TrackSourceMicrophone, 48000, 2, 48)
	if err != nil {
		t.Fatalf("failed to publish audio track: %v", err)
	}
	if track.Codec != CodecOpus || track.BitrateKbps != 48 || track.Channels != 2 {
		t.Errorf("unexpected track properties: %+v", track)
	}

	var speakingState bool
	vc.OnSpeaking(func(userID string, speaking bool) {
		speakingState = speaking
	})
	_ = vc.SetSpeaking(true)
	if !speakingState {
		t.Errorf("expected speaking state to be true")
	}
}

func TestInteractionSerialization(t *testing.T) {
	cmd := &ApplicationCommand{
		Name:        "ping",
		Description: "A slash command",
		Type:        ChatInputCommandType,
		Options: []*ApplicationCommandOption{
			{
				Type:        OptionTypeString,
				Name:        "message",
				Description: "Text to send",
				Required:    true,
			},
		},
	}

	data, err := json.Marshal(cmd)
	if err != nil {
		t.Fatalf("failed to marshal application command: %v", err)
	}

	var parsed ApplicationCommand
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal application command: %v", err)
	}

	if parsed.Name != "ping" || len(parsed.Options) != 1 {
		t.Errorf("unexpected command structure: %+v", parsed)
	}
}

func TestShardingManager(t *testing.T) {
	mgr, err := NewShardingManager("test-token", 2)
	if err != nil {
		t.Fatalf("failed to create sharding manager: %v", err)
	}

	if len(mgr.Sessions) != 2 {
		t.Errorf("expected 2 shard sessions, got %d", len(mgr.Sessions))
	}
	if mgr.Sessions[0].options.ShardID != 0 || mgr.Sessions[1].options.ShardID != 1 {
		t.Errorf("shard IDs incorrect: %d, %d", mgr.Sessions[0].options.ShardID, mgr.Sessions[1].options.ShardID)
	}
}

func TestAudioStreamer(t *testing.T) {
	streamer := NewAudioStreamer(nil)
	defer streamer.Close()

	frame := []byte{0xF8, 0xFF, 0xFE}
	err := streamer.SendOpusFrame(frame)
	if err != nil {
		t.Errorf("expected frame to send, got: %v", err)
	}
}

func TestStateCache(t *testing.T) {
	state := NewState()
	state.MaxMessageHistory = 2

	// Test Ready
	state.OnReady(&Ready{
		GatewayReady: &GatewayReady{
			User: &User{ID: "bot_1", Username: "TestBot"},
		},
	})
	if state.GetUser() == nil || state.GetUser().Username != "TestBot" {
		t.Errorf("expected bot user to be cached")
	}

	// Test Guild Create with channels and members
	state.OnGuildCreate(&GuildCreate{
		Guild: &Guild{
			ID:   "g_1",
			Name: "Test Server",
			Channels: []*Channel{
				{ID: "c_1", Name: "general"},
			},
			Roles: []*Role{
				{ID: "r_1", Name: "Admin"},
			},
			Members: []*Member{
				{User: &User{ID: "u_1", Username: "Alice"}},
			},
		},
	})

	g, err := state.Guild("g_1")
	if err != nil || g.Name != "Test Server" {
		t.Errorf("expected guild in state, got %v", err)
	}

	ch, err := state.Channel("c_1")
	if err != nil || ch.Name != "general" {
		t.Errorf("expected channel in state, got %v", err)
	}

	m, err := state.Member("g_1", "u_1")
	if err != nil || m.User.Username != "Alice" {
		t.Errorf("expected member in state, got %v", err)
	}

	r, err := state.Role("g_1", "r_1")
	if err != nil || r.Name != "Admin" {
		t.Errorf("expected role in state, got %v", err)
	}

	// Test Message Ring Buffer
	state.OnMessageCreate(&MessageCreate{Message: &Message{ID: "m1", ChannelID: "c_1", Content: "First"}})
	state.OnMessageCreate(&MessageCreate{Message: &Message{ID: "m2", ChannelID: "c_1", Content: "Second"}})
	state.OnMessageCreate(&MessageCreate{Message: &Message{ID: "m3", ChannelID: "c_1", Content: "Third"}})

	msgs := state.MessagesList("c_1")
	if len(msgs) != 2 {
		t.Errorf("expected max 2 messages in ring buffer, got %d", len(msgs))
	}
	if msgs[0].ID != "m2" || msgs[1].ID != "m3" {
		t.Errorf("expected m2 and m3 in buffer, got %s and %s", msgs[0].ID, msgs[1].ID)
	}

	// Test Guild Delete
	state.OnGuildDelete(&GuildDelete{ID: "g_1"})
	if _, err := state.Guild("g_1"); err == nil {
		t.Errorf("expected guild to be deleted from state")
	}
	if _, err := state.Channel("c_1"); err == nil {
		t.Errorf("expected channel to be deleted from state on guild delete")
	}
}

func TestPresenceHelpers(t *testing.T) {
	game := ActivityGame("Micup Arena")
	if game.Type != ActivityTypeGame || game.Name != "Micup Arena" {
		t.Errorf("unexpected game activity: %+v", game)
	}

	stream := ActivityStreaming("Live Coding", "https://twitch.tv/example")
	if stream.Type != ActivityTypeStreaming || stream.URL != "https://twitch.tv/example" {
		t.Errorf("unexpected streaming activity: %+v", stream)
	}

	custom := ActivityCustom("Developing bots...")
	if custom.Type != ActivityTypeCustom || custom.State != "Developing bots..." {
		t.Errorf("unexpected custom activity: %+v", custom)
	}
}

func TestComponentRouter(t *testing.T) {
	router := NewComponentRouter()
	var handled atomic.Bool

	router.Handle("btn_test", func(s *Session, i *Interaction) {
		handled.Store(true)
	})

	listener := router.Listener()
	listener(nil, &InteractionCreate{
		Interaction: &Interaction{
			ID: "int_1",
			Data: &InteractionData{
				CustomID: "btn_test",
			},
		},
	})

	if !handled.Load() {
		t.Errorf("expected component handler to be invoked")
	}
}

func TestWebhookSerialization(t *testing.T) {
	params := &WebhookExecuteParams{
		Content:   "Hello from Webhook!",
		Username:  "Webhooker",
		AvatarURL: "https://example.com/avatar.png",
		Embeds: []*Embed{
			NewEmbed().SetTitle("Test Webhook Embed").Build(),
		},
	}

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("failed to marshal webhook params: %v", err)
	}

	var parsed WebhookExecuteParams
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal webhook params: %v", err)
	}

	if parsed.Username != "Webhooker" || len(parsed.Embeds) != 1 {
		t.Errorf("unexpected parsed webhook params: %+v", parsed)
	}
}

func TestPermissionsBitwiseAndOverwrites(t *testing.T) {
	// Parse and format
	permStr := "1048576" // PermissionConnect (1 << 20)
	permInt := ParsePermissions(permStr)
	if permInt != PermissionConnect {
		t.Errorf("expected %d, got %d", PermissionConnect, permInt)
	}
	if FormatPermissions(permInt) != permStr {
		t.Errorf("expected string %s, got %s", permStr, FormatPermissions(permInt))
	}

	// Bitwise operations
	base := PermissionSendMessages | PermissionViewChannel
	if !HasPermission(base, PermissionSendMessages) {
		t.Errorf("expected base to have SendMessages")
	}
	if HasPermission(base, PermissionAdministrator) {
		t.Errorf("did not expect base to have Administrator")
	}

	// Administrator override
	adminPerms := PermissionAdministrator
	if !HasPermission(adminPerms, PermissionManageGuild) {
		t.Errorf("Administrator should have all permissions")
	}

	// ComputeBasePermissions & Overwrites
	guild := &Guild{
		ID:      "guild_1",
		OwnerID: "owner_1",
		Roles: []*Role{
			{ID: "guild_1", Permissions: FormatPermissions(PermissionViewChannel)},                               // @everyone
			{ID: "role_mod", Permissions: FormatPermissions(PermissionSendMessages | PermissionManageMessages)}, // Mod role
		},
	}

	member := &Member{
		User:  &User{ID: "member_1"},
		Roles: []string{"role_mod"},
	}

	calculatedBase := ComputeBasePermissions(member, guild)
	if !HasPermission(calculatedBase, PermissionSendMessages) || !HasPermission(calculatedBase, PermissionViewChannel) {
		t.Errorf("expected member to have base SendMessages and ViewChannel, got %d", calculatedBase)
	}

	// Channel overwrites
	channel := &Channel{
		ID:      "chan_1",
		GuildID: "guild_1",
		PermissionOverwrites: []*PermissionOverwrite{
			{
				ID:    "role_mod",
				Type:  0, // Role
				Deny:  FormatPermissions(PermissionSendMessages),
				Allow: "0",
			},
			{
				ID:    "member_1",
				Type:  1, // Member override
				Allow: FormatPermissions(PermissionSendMessages),
				Deny:  "0",
			},
		},
	}

	effective := ComputeOverwrites(calculatedBase, member, channel)
	if !HasPermission(effective, PermissionSendMessages) {
		t.Errorf("expected member-specific allow overwrite to restore SendMessages permission")
	}
}

func TestAuditLogSerializationAndQuery(t *testing.T) {
	entry := &AuditLogEntry{
		ID:         "entry_1",
		TargetID:   "user_target",
		UserID:     "user_mod",
		ActionType: AuditLogActionMemberBanAdd,
		Reason:     "Spamming invites",
		Changes: []*AuditLogChange{
			{
				Key:      "communication_disabled_until",
				NewValue: "2026-11-08T12:00:00Z",
			},
		},
		Options: &AuditLogOptionsData{
			DeleteMemberDays: "1",
		},
	}

	logData := &AuditLog{
		Entries: []*AuditLogEntry{entry},
		Users: []*User{
			{ID: "user_mod", Username: "ModeratorBot"},
		},
	}

	data, err := json.Marshal(logData)
	if err != nil {
		t.Fatalf("failed to marshal audit log: %v", err)
	}

	var parsed AuditLog
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal audit log: %v", err)
	}

	if len(parsed.Entries) != 1 || parsed.Entries[0].ActionType != AuditLogActionMemberBanAdd {
		t.Errorf("unexpected parsed audit log entries: %+v", parsed.Entries)
	}
	if parsed.Entries[0].Options.DeleteMemberDays != "1" {
		t.Errorf("expected DeleteMemberDays to be 1, got %s", parsed.Entries[0].Options.DeleteMemberDays)
	}
}

func TestEmojiAndReactionPayloads(t *testing.T) {
	createParams := &CreateEmojiParams{
		Name:  "custom_cat",
		Image: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
		Roles: []string{"role_vip"},
	}

	data, err := json.Marshal(createParams)
	if err != nil {
		t.Fatalf("failed to marshal create emoji params: %v", err)
	}

	var parsed CreateEmojiParams
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal create emoji params: %v", err)
	}

	if parsed.Name != "custom_cat" || len(parsed.Roles) != 1 {
		t.Errorf("unexpected parsed emoji params: %+v", parsed)
	}
}

func TestScheduledEventModels(t *testing.T) {
	now := time.Now()
	endTime := now.Add(2 * time.Hour)

	event := &GuildScheduledEvent{
		ID:                 "event_1",
		GuildID:            "guild_1",
		Name:               "Community Game Night",
		Description:        "Join us for tournament play!",
		ScheduledStartTime: now,
		ScheduledEndTime:   &endTime,
		PrivacyLevel:       EventPrivacyGuildOnly,
		Status:             EventStatusScheduled,
		EntityType:         EventEntityExternal,
		EntityMetadata: &ScheduledEventEntityMetadata{
			Location: "https://micup.gg/tournament",
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal scheduled event: %v", err)
	}

	var parsed GuildScheduledEvent
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal scheduled event: %v", err)
	}

	if parsed.Name != "Community Game Night" || parsed.EntityMetadata.Location != "https://micup.gg/tournament" {
		t.Errorf("unexpected parsed event: %+v", parsed)
	}
	if parsed.Status != EventStatusScheduled || parsed.EntityType != EventEntityExternal {
		t.Errorf("unexpected event status/type: status=%d type=%d", parsed.Status, parsed.EntityType)
	}
}

func TestAutoModEngine(t *testing.T) {
	engine := NewAutoModEngine()

	engine.AddKeywordRule("anti_badword", []string{"badword", "nastyterm"}, nil)
	_ = engine.AddRegexRule("anti_invite", []string{`(?i)micup\.gg/invite/[a-z0-9]+`}, nil)
	engine.AddMentionLimitRule("anti_mass_mention", 3, nil)

	// Clean message
	cleanMsg := &Message{Content: "Hello world!", Mentions: []*User{{ID: "u1"}}}
	matched, _, _ := engine.CheckMessage(cleanMsg)
	if matched {
		t.Errorf("expected clean message not to trigger automod")
	}

	// Keyword match
	badMsg := &Message{Content: "Hey check this badword out"}
	matched, rule, _ := engine.CheckMessage(badMsg)
	if !matched || rule != "anti_badword" {
		t.Errorf("expected anti_badword match, got matched=%v rule=%s", matched, rule)
	}

	// Regex match
	inviteMsg := &Message{Content: "Join my cool guild: micup.gg/invite/xyz123"}
	matched, rule, _ = engine.CheckMessage(inviteMsg)
	if !matched || rule != "anti_invite" {
		t.Errorf("expected anti_invite match, got matched=%v rule=%s", matched, rule)
	}

	// Mass mention match
	spamMsg := &Message{
		Content: "Ping all",
		Mentions: []*User{
			{ID: "u1"}, {ID: "u2"}, {ID: "u3"}, {ID: "u4"},
		},
	}
	matched, rule, _ = engine.CheckMessage(spamMsg)
	if !matched || rule != "anti_mass_mention" {
		t.Errorf("expected anti_mass_mention match, got matched=%v rule=%s", matched, rule)
	}
}

func TestAudioPlayerVolumeScaling(t *testing.T) {
	// 2 samples: 1000 and -1000 in little endian 16-bit
	pcm := make([]byte, 4)
	valPos := int16(1000)
	valNeg := int16(-1000)
	binary.LittleEndian.PutUint16(pcm[0:2], uint16(valPos))
	binary.LittleEndian.PutUint16(pcm[2:4], uint16(valNeg))

	// Scale 50%
	scaled50 := ScalePCM16(pcm, 0.5)
	s1 := int16(binary.LittleEndian.Uint16(scaled50[0:2]))
	s2 := int16(binary.LittleEndian.Uint16(scaled50[2:4]))
	if s1 != 500 || s2 != -500 {
		t.Errorf("expected 500 and -500, got %d and %d", s1, s2)
	}

	// Scale 0% (silence)
	scaled0 := ScalePCM16(pcm, 0.0)
	z1 := int16(binary.LittleEndian.Uint16(scaled0[0:2]))
	z2 := int16(binary.LittleEndian.Uint16(scaled0[2:4]))
	if z1 != 0 || z2 != 0 {
		t.Errorf("expected silence (0), got %d and %d", z1, z2)
	}

	// Boundary clamping
	loudPCM := make([]byte, 4)
	loudPos := int16(20000)
	loudNeg := int16(-20000)
	binary.LittleEndian.PutUint16(loudPCM[0:2], uint16(loudPos))
	binary.LittleEndian.PutUint16(loudPCM[2:4], uint16(loudNeg))
	clamped := ScalePCM16(loudPCM, 2.0)
	c1 := int16(binary.LittleEndian.Uint16(clamped[0:2]))
	c2 := int16(binary.LittleEndian.Uint16(clamped[2:4]))
	if c1 != 32767 || c2 != -32768 {
		t.Errorf("expected boundary clamp (32767, -32768), got %d and %d", c1, c2)
	}
}

func TestHardwareEncoderProbe(t *testing.T) {
	cap := ProbeHardwareEncoders()
	if cap == nil {
		t.Fatalf("expected non-nil hardware encoder capability")
	}
	if len(cap.SupportedCodecs) == 0 {
		t.Errorf("expected at least one supported codec")
	}
	if cap.PreferredCodec == "" {
		t.Errorf("expected non-empty preferred codec")
	}
}

func TestGameActivityValidationAndPayload(t *testing.T) {
	// 1. Game ID validation
	if !ValidateGameID("counter-strike-2") || !ValidateGameID("valorant") {
		t.Errorf("expected valid game IDs to pass")
	}
	if ValidateGameID("Invalid_Caps") || ValidateGameID("-dash-start") || ValidateGameID("toolongggggggggggggggggggggggggggggggggggggggggggggg") {
		t.Errorf("expected invalid game IDs to fail")
	}

	// 2. Manual Name validation
	if !ValidateManualName("Minecraft 1.20") || !ValidateManualName("Custom Game") {
		t.Errorf("expected valid manual names to pass")
	}
	if ValidateManualName("") || ValidateManualName("   ") || ValidateManualName("Bad\x00Control") {
		t.Errorf("expected invalid manual names to fail")
	}

	// 3. Executable validation
	if !ValidateExecutable("cs2.exe") || !ValidateExecutable("game.exe") {
		t.Errorf("expected valid exe to pass")
	}
	if ValidateExecutable("javaw.exe") || ValidateExecutable("steam.exe") || ValidateExecutable("game.sh") || ValidateExecutable("../path/game.exe") {
		t.Errorf("expected blacklisted or invalid exe to fail")
	}

	// 4. Payload generation & serialization
	now := time.Now()
	payload, err := NewGameActivityPayload("valorant", "", now)
	if err != nil {
		t.Fatalf("unexpected error creating game activity payload: %v", err)
	}

	if payload.State != GameActivityStatePlaying || payload.Activity.GameID != "valorant" {
		t.Errorf("unexpected game activity state or game_id: %+v", payload)
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal game activity payload: %v", err)
	}

	var parsed GameActivityPayload
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal game activity payload: %v", err)
	}
	if parsed.Activity.StartedAt != now.UnixMilli() {
		t.Errorf("expected started_at %d, got %d", now.UnixMilli(), parsed.Activity.StartedAt)
	}
}

func TestVideoTrackAndStreamer(t *testing.T) {
	vc := &VoiceConnection{
		UserID:   "bot_1",
		Streamer: NewAudioStreamer(nil),
	}

	// Publish video track
	track, err := vc.PublishVideoTrack(TrackSourceCamera, CodecH264, 1920, 1080, 60, 4500)
	if err != nil {
		t.Fatalf("failed to publish video track: %v", err)
	}

	if track.Kind != TrackKindVideo || track.Codec != CodecH264 || track.Width != 1920 || track.Height != 1080 || track.FPS != 60 {
		t.Errorf("unexpected video track properties: %+v", track)
	}

	// Queue frame
	frameData := []byte{0x00, 0x00, 0x00, 0x01, 0x67, 0x42, 0x00, 0x0a}
	if err := vc.SendVideoFrame(frameData); err != nil {
		t.Errorf("failed to send video frame: %v", err)
	}

	// Close video streamer
	if err := vc.VideoStreamer.Close(); err != nil {
		t.Errorf("failed to close video streamer: %v", err)
	}
}


