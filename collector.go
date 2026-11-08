package micup

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrCollectorTimeout   = errors.New("collector timed out waiting for event")
	ErrCollectorCancelled = errors.New("collector context cancelled")
)

// WaitForMessage waits for a message that matches the provided filter function, or context expiration.
func (s *Session) WaitForMessage(ctx context.Context, filter func(*Message) bool) (*Message, error) {
	ch := make(chan *Message, 1)

	unregister := s.AddHandler(func(m *MessageCreate) {
		if filter == nil || filter(m.Message) {
			select {
			case ch <- m.Message:
			default:
			}
		}
	})
	defer unregister()

	select {
	case msg := <-ch:
		return msg, nil
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, ErrCollectorTimeout
		}
		return nil, ctx.Err()
	}
}

// WaitForInteraction waits for an interaction with the given custom_id.
func (s *Session) WaitForInteraction(ctx context.Context, customID string) (*Interaction, error) {
	ch := make(chan *Interaction, 1)

	unregister := s.AddHandler(func(i *InteractionCreate) {
		if i.Interaction != nil && i.Data != nil && i.Data.CustomID == customID {
			select {
			case ch <- i.Interaction:
			default:
			}
		}
	})
	defer unregister()

	select {
	case interaction := <-ch:
		return interaction, nil
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, ErrCollectorTimeout
		}
		return nil, ctx.Err()
	}
}

// ComponentHandler represents a callback for interactive button/select component clicks.
type ComponentHandler func(s *Session, i *Interaction)

// ComponentRouter routes UI button / select menu interactions by their custom_id.
type ComponentRouter struct {
	mu       sync.RWMutex
	handlers map[string]ComponentHandler
}

// NewComponentRouter creates a new component router.
func NewComponentRouter() *ComponentRouter {
	return &ComponentRouter{
		handlers: make(map[string]ComponentHandler),
	}
}

// Handle registers a handler function for a specific component custom_id.
func (cr *ComponentRouter) Handle(customID string, handler ComponentHandler) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.handlers[customID] = handler
}

// Remove unregisters a component handler.
func (cr *ComponentRouter) Remove(customID string) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	delete(cr.handlers, customID)
}

// Listener returns an event handler function suitable for Session.AddHandler.
func (cr *ComponentRouter) Listener() func(s *Session, i *InteractionCreate) {
	return func(s *Session, i *InteractionCreate) {
		if i == nil || i.Interaction == nil || i.Data == nil {
			return
		}

		cr.mu.RLock()
		handler, exists := cr.handlers[i.Data.CustomID]
		cr.mu.RUnlock()

		if exists && handler != nil {
			handler(s, i.Interaction)
		}
	}
}

// EmbedPaginator handles interactive multi-page embed cards with Next and Previous buttons.
type EmbedPaginator struct {
	Session   *Session
	ChannelID string
	Pages     []*Embed
	Current   int
	Timeout   time.Duration
	PrevID    string
	NextID    string
}

// NewEmbedPaginator creates a new paginator for multiple embed pages.
func NewEmbedPaginator(s *Session, channelID string, pages []*Embed) *EmbedPaginator {
	return &EmbedPaginator{
		Session:   s,
		ChannelID: channelID,
		Pages:     pages,
		Current:   0,
		Timeout:   2 * time.Minute,
		PrevID:    "paginator_prev",
		NextID:    "paginator_next",
	}
}

// Send sends the first page with interactive Next/Previous buttons and starts the collector.
func (p *EmbedPaginator) Send(ctx context.Context) (*Message, error) {
	if len(p.Pages) == 0 {
		return nil, errors.New("paginator has no pages")
	}

	buttons := []Component{
		NewButton(ButtonStyleSecondary, p.PrevID, "◀ Önceki").WithDisabled(p.Current == 0),
		NewButton(ButtonStylePrimary, p.NextID, "Sonraki ▶").WithDisabled(p.Current >= len(p.Pages)-1),
	}

	msgData := &MessageSend{
		Embeds: []*Embed{p.Pages[p.Current]},
		Components: []Component{
			NewActionRow(buttons...),
		},
	}

	return p.Session.ChannelMessageSendComplex(p.ChannelID, msgData)
}
