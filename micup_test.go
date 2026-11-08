package micup

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

	received := false
	unregister := registry.Add(func(m *MessageCreate) {
		if m.Content == "ping" {
			received = true
		}
	})

	registry.Dispatch(nil, &MessageCreate{
		Message: &Message{Content: "ping"},
	})

	time.Sleep(50 * time.Millisecond)
	if !received {
		t.Error("expected event to be received by handler")
	}

	// Test unregister
	unregister()
	received = false
	registry.Dispatch(nil, &MessageCreate{
		Message: &Message{Content: "ping"},
	})
	time.Sleep(50 * time.Millisecond)
	if received {
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
