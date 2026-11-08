package micup

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
)

// Event structs wrapping dispatch event payloads.
type Ready struct {
	*GatewayReady
}

type Resumed struct{}

type MessageCreate struct {
	*Message
}

type MessageUpdate struct {
	*Message
}

type MessageDelete struct {
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
	GuildID   string `json:"guild_id,omitempty"`
}

type MessageBulkDelete struct {
	IDs       []string `json:"ids"`
	ChannelID string   `json:"channel_id"`
	GuildID   string   `json:"guild_id,omitempty"`
}

type GuildCreate struct {
	*Guild
}

type GuildUpdate struct {
	*Guild
}

type GuildDelete struct {
	ID          string `json:"id"`
	Unavailable bool   `json:"unavailable,omitempty"`
}

type GuildMemberAdd struct {
	*Member
	GuildID string `json:"guild_id"`
}

type GuildMemberRemove struct {
	GuildID string `json:"guild_id"`
	User    *User  `json:"user"`
}

type GuildMemberUpdate struct {
	GuildID string   `json:"guild_id"`
	Roles   []string `json:"roles"`
	User    *User    `json:"user"`
	Nick    string   `json:"nick,omitempty"`
}

type ChannelCreate struct {
	*Channel
}

type ChannelUpdate struct {
	*Channel
}

type ChannelDelete struct {
	*Channel
}

type ChannelPinsUpdate struct {
	ChannelID        string `json:"channel_id"`
	GuildID          string `json:"guild_id,omitempty"`
	LastPinTimestamp string `json:"last_pin_timestamp,omitempty"`
}

type TypingStart struct {
	ChannelID string  `json:"channel_id"`
	GuildID   string  `json:"guild_id,omitempty"`
	UserID    string  `json:"user_id"`
	Timestamp int64   `json:"timestamp"`
	Member    *Member `json:"member,omitempty"`
}

type UserUpdate struct {
	*User
}

type VoiceStateUpdate struct {
	GuildID   string `json:"guild_id,omitempty"`
	ChannelID string `json:"channel_id,omitempty"`
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	Deaf      bool   `json:"deaf"`
	Mute      bool   `json:"mute"`
	SelfDeaf  bool   `json:"self_deaf"`
	SelfMute  bool   `json:"self_mute"`
}

type VoiceServerUpdate struct {
	Token    string `json:"token"`
	GuildID  string `json:"guild_id"`
	Endpoint string `json:"endpoint"`
}

type InteractionCreate struct {
	*Interaction
}

// EventHandlerRegistry handles registration and dispatch of event callbacks.
type EventHandlerRegistry struct {
	mu       sync.RWMutex
	handlers []reflect.Value
}

// NewEventHandlerRegistry creates a new event handler registry.
func NewEventHandlerRegistry() *EventHandlerRegistry {
	return &EventHandlerRegistry{
		handlers: make([]reflect.Value, 0),
	}
}

// Add adds an event listener function.
// Handlers can have signatures:
//   func(s *Session, e *EventType)
//   func(e *EventType)
func (r *EventHandlerRegistry) Add(handler interface{}) func() {
	v := reflect.ValueOf(handler)
	if v.Kind() != reflect.Func {
		panic("event handler must be a function")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.handlers = append(r.handlers, v)

	// Return unregister func
	return func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		for i, h := range r.handlers {
			if h == v {
				r.handlers = append(r.handlers[:i], r.handlers[i+1:]...)
				break
			}
		}
	}
}

// Dispatch routes an event object to all matching registered handlers.
func (r *EventHandlerRegistry) Dispatch(session *Session, event interface{}) {
	r.mu.RLock()
	handlers := make([]reflect.Value, len(r.handlers))
	copy(handlers, r.handlers)
	r.mu.RUnlock()

	eventVal := reflect.ValueOf(event)
	eventType := eventVal.Type()
	sessionVal := reflect.ValueOf(session)

	for _, handler := range handlers {
		hType := handler.Type()
		numIn := hType.NumIn()

		if numIn == 1 && hType.In(0) == eventType {
			go handler.Call([]reflect.Value{eventVal})
		} else if numIn == 2 && hType.In(0) == sessionVal.Type() && hType.In(1) == eventType {
			go handler.Call([]reflect.Value{sessionVal, eventVal})
		}
	}
}

// parseDispatchEvent deserializes JSON into the corresponding event struct.
func parseDispatchEvent(eventName string, raw []byte) (interface{}, error) {
	switch eventName {
	case EventReady:
		var ready GatewayReady
		if err := json.Unmarshal(raw, &ready); err != nil {
			return nil, err
		}
		return &Ready{GatewayReady: &ready}, nil

	case EventResumed:
		return &Resumed{}, nil

	case EventMessageCreate:
		var msg Message
		if err := json.Unmarshal(raw, &msg); err != nil {
			return nil, err
		}
		return &MessageCreate{Message: &msg}, nil

	case EventMessageUpdate:
		var msg Message
		if err := json.Unmarshal(raw, &msg); err != nil {
			return nil, err
		}
		return &MessageUpdate{Message: &msg}, nil

	case EventMessageDelete:
		var del MessageDelete
		if err := json.Unmarshal(raw, &del); err != nil {
			return nil, err
		}
		return &del, nil

	case EventMessageBulkDelete:
		var bulk MessageBulkDelete
		if err := json.Unmarshal(raw, &bulk); err != nil {
			return nil, err
		}
		return &bulk, nil

	case EventGuildCreate:
		var guild Guild
		if err := json.Unmarshal(raw, &guild); err != nil {
			return nil, err
		}
		return &GuildCreate{Guild: &guild}, nil

	case EventGuildUpdate:
		var guild Guild
		if err := json.Unmarshal(raw, &guild); err != nil {
			return nil, err
		}
		return &GuildUpdate{Guild: &guild}, nil

	case EventGuildDelete:
		var del GuildDelete
		if err := json.Unmarshal(raw, &del); err != nil {
			return nil, err
		}
		return &del, nil

	case EventGuildMemberAdd:
		var member GuildMemberAdd
		if err := json.Unmarshal(raw, &member); err != nil {
			return nil, err
		}
		return &member, nil

	case EventGuildMemberRemove:
		var rem GuildMemberRemove
		if err := json.Unmarshal(raw, &rem); err != nil {
			return nil, err
		}
		return &rem, nil

	case EventGuildMemberUpdate:
		var upd GuildMemberUpdate
		if err := json.Unmarshal(raw, &upd); err != nil {
			return nil, err
		}
		return &upd, nil

	case EventChannelCreate:
		var ch Channel
		if err := json.Unmarshal(raw, &ch); err != nil {
			return nil, err
		}
		return &ChannelCreate{Channel: &ch}, nil

	case EventChannelUpdate:
		var ch Channel
		if err := json.Unmarshal(raw, &ch); err != nil {
			return nil, err
		}
		return &ChannelUpdate{Channel: &ch}, nil

	case EventChannelDelete:
		var ch Channel
		if err := json.Unmarshal(raw, &ch); err != nil {
			return nil, err
		}
		return &ChannelDelete{Channel: &ch}, nil

	case EventChannelPinsUpdate:
		var pins ChannelPinsUpdate
		if err := json.Unmarshal(raw, &pins); err != nil {
			return nil, err
		}
		return &pins, nil

	case EventTypingStart:
		var typing TypingStart
		if err := json.Unmarshal(raw, &typing); err != nil {
			return nil, err
		}
		return &typing, nil

	case EventUserUpdate:
		var user User
		if err := json.Unmarshal(raw, &user); err != nil {
			return nil, err
		}
		return &UserUpdate{User: &user}, nil

	case EventVoiceStateUpdate:
		var voice VoiceStateUpdate
		if err := json.Unmarshal(raw, &voice); err != nil {
			return nil, err
		}
		return &voice, nil

	case EventVoiceServerUpdate:
		var vServer VoiceServerUpdate
		if err := json.Unmarshal(raw, &vServer); err != nil {
			return nil, err
		}
		return &vServer, nil

	case EventInteractionCreate:
		var inter Interaction
		if err := json.Unmarshal(raw, &inter); err != nil {
			return nil, err
		}
		return &InteractionCreate{Interaction: &inter}, nil

	default:
		return nil, fmt.Errorf("unknown dispatch event type: %s", eventName)
	}
}
