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
