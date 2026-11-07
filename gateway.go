package micup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// GatewayClient manages the live WebSocket connection to Micup.
type GatewayClient struct {
	session         *Session
	conn            *websocket.Conn
	mu              sync.Mutex
	writeMu         sync.Mutex
	status          int32 // 0 = disconnected, 1 = connected, 2 = connecting
	token           string
	gatewayURL      string
	sessionID       string
	lastSequence    int64
	heartbeatStop   chan struct{}
	lastHeartbeat   time.Time
	lastAck         time.Time
	heartbeatAcked  bool
	reconnectBackoff time.Duration
	ctx             context.Context
	cancel          context.CancelFunc
}

// Connection states
const (
	statusDisconnected = 0
	statusConnected    = 1
	statusConnecting   = 2
)

// NewGatewayClient creates a new gateway client.
func NewGatewayClient(session *Session, token string) *GatewayClient {
	return &GatewayClient{
		session:          session,
		token:            token,
		reconnectBackoff: 1 * time.Second,
	}
}

// Open initiates the gateway connection and starts the message reader loop.
func (g *GatewayClient) Open(ctx context.Context) error {
	g.mu.Lock()
	if atomic.LoadInt32(&g.status) != statusDisconnected {
		g.mu.Unlock()
		return errors.New("gateway is already connected or connecting")
	}
	atomic.StoreInt32(&g.status, statusConnecting)
	g.ctx, g.cancel = context.WithCancel(ctx)
	g.mu.Unlock()

	return g.connect()
}

// Close gracefully terminates the gateway connection.
func (g *GatewayClient) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if atomic.LoadInt32(&g.status) == statusDisconnected {
		return nil
	}

	if g.cancel != nil {
		g.cancel()
	}

	g.stopHeartbeat()

	if g.conn != nil {
		_ = g.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(CloseNormalClosure, "Client shutting down"))
		_ = g.conn.Close()
		g.conn = nil
	}

	atomic.StoreInt32(&g.status, statusDisconnected)
	return nil
}

// resolveGatewayURL fetches the gateway URL from discovery if not set.
func (g *GatewayClient) resolveGatewayURL() (string, error) {
	if g.gatewayURL != "" {
		return g.gatewayURL, nil
	}

	if g.session.discovery != nil {
		eps, err := g.session.discovery.GetEndpoints(g.ctx)
		if err == nil && eps.Gateway != "" {
			g.gatewayURL = fmt.Sprintf("%s?v=1&encoding=json", eps.Gateway)
			return g.gatewayURL, nil
		}
	}

	g.gatewayURL = "wss://micup.gg/gateway?v=1&encoding=json"
	return g.gatewayURL, nil
}

// connect establishes the websocket connection.
func (g *GatewayClient) connect() error {
	gwURL, err := g.resolveGatewayURL()
	if err != nil {
		atomic.StoreInt32(&g.status, statusDisconnected)
		return err
	}

	header := http.Header{}
	header.Set("User-Agent", DefaultUserAgent)

	dialer := websocket.Dialer{
		HandshakeTimeout: 15 * time.Second,
	}

	conn, _, err := dialer.DialContext(g.ctx, gwURL, header)
	if err != nil {
		atomic.StoreInt32(&g.status, statusDisconnected)
		return fmt.Errorf("failed to dial gateway websocket: %w", err)
	}

	g.mu.Lock()
	g.conn = conn
	atomic.StoreInt32(&g.status, statusConnected)
	g.mu.Unlock()

	go g.readLoop()
	return nil
}

// readLoop reads and routes incoming WebSocket messages.
func (g *GatewayClient) readLoop() {
	defer func() {
		g.handleDisconnect()
	}()

	for {
		select {
		case <-g.ctx.Done():
			return
		default:
		}

		g.mu.Lock()
		conn := g.conn
		g.mu.Unlock()

		if conn == nil {
			return
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[micup-gateway] websocket read error: %v", err)
			return
		}

		var payload GatewayPayload
		if err := json.Unmarshal(message, &payload); err != nil {
			log.Printf("[micup-gateway] payload json unmarshal error: %v", err)
			continue
		}

		g.routePayload(&payload)
	}
}

// routePayload processes opcodes from the server.
func (g *GatewayClient) routePayload(payload *GatewayPayload) {
	switch payload.Op {
	case OpHello:
		var hello GatewayHello
		if err := json.Unmarshal(payload.D, &hello); err != nil {
			log.Printf("[micup-gateway] failed to parse OpHello: %v", err)
			return
		}
		g.startHeartbeat(hello.HeartbeatInterval)
		if g.canResume() {
			g.sendResume()
		} else {
			g.sendIdentify()
		}

	case OpHeartbeat:
		// Server requested immediate heartbeat
		g.sendHeartbeat()

	case OpHeartbeatAck:
		g.lastAck = time.Now()
		g.heartbeatAcked = true

	case OpReconnect:
		log.Println("[micup-gateway] server requested reconnect")
		g.disconnectAndReconnect(false)

	case OpInvalidSession:
		var resumable bool
		_ = json.Unmarshal(payload.D, &resumable)
		log.Printf("[micup-gateway] invalid session (resumable=%v)", resumable)
		if !resumable {
			g.sessionID = ""
			g.lastSequence = 0
		}
		// Delay briefly before reconnecting
		time.Sleep(time.Duration(1000+rand.Intn(2000)) * time.Millisecond)
		g.disconnectAndReconnect(false)

	case OpGatewayError:
		var gErr GatewayError
		_ = json.Unmarshal(payload.D, &gErr)
		log.Printf("[micup-gateway] gateway error received [%s]: %s", gErr.Code, gErr.Message)

	case OpDispatch:
		if payload.S != nil {
			atomic.StoreInt64(&g.lastSequence, *payload.S)
		}

		if payload.T == EventReady {
			var ready GatewayReady
			if err := json.Unmarshal(payload.D, &ready); err == nil {
				g.sessionID = ready.SessionID
				g.reconnectBackoff = 1 * time.Second // reset backoff
				log.Printf("[micup-gateway] connected to Micup! User: %s (ID: %s), Session: %s",
					ready.User.Username, ready.User.ID, ready.SessionID)
			}
		} else if payload.T == EventResumed {
			g.reconnectBackoff = 1 * time.Second
			log.Printf("[micup-gateway] session %s successfully resumed", g.sessionID)
		}

		// Parse and dispatch typed event
		event, err := parseDispatchEvent(payload.T, payload.D)
		if err == nil && event != nil {
			g.session.events.Dispatch(g.session, event)
		}
	}
}

// canResume checks if a valid session can be resumed.
func (g *GatewayClient) canResume() bool {
	return g.sessionID != "" && atomic.LoadInt64(&g.lastSequence) > 0
}

// sendIdentify sends OpIdentify to establish new session.
func (g *GatewayClient) sendIdentify() {
	identify := GatewayIdentify{
		Token: g.token,
		Properties: GatewayIdentifyProperties{
			OS:      runtime.GOOS,
			Browser: "micup-go",
			Device:  "micup-go",
		},
		Presence: &GatewayPresence{
			Status: "online",
			AFK:    false,
		},
	}

	data, _ := json.Marshal(identify)
	_ = g.SendPayload(&GatewayPayload{
		Op: OpIdentify,
		D:  data,
	})
}

// sendResume sends OpResume to continue existing session.
func (g *GatewayClient) sendResume() {
	resume := GatewayResume{
		Token:     g.token,
		SessionID: g.sessionID,
		Seq:       atomic.LoadInt64(&g.lastSequence),
	}

	data, _ := json.Marshal(resume)
	_ = g.SendPayload(&GatewayPayload{
		Op: OpResume,
		D:  data,
	})
}

// startHeartbeat begins the periodic heartbeat ping.
func (g *GatewayClient) startHeartbeat(intervalMs int) {
	g.stopHeartbeat()

	g.heartbeatStop = make(chan struct{})
	g.heartbeatAcked = true

	go func() {
		// Initial jitter: interval * rand.Float64()
		initialDelay := time.Duration(float64(intervalMs)*rand.Float64()) * time.Millisecond
		select {
		case <-time.After(initialDelay):
			g.sendHeartbeat()
		case <-g.heartbeatStop:
			return
		case <-g.ctx.Done():
			return
		}

		ticker := time.NewTicker(time.Duration(intervalMs) * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if !g.heartbeatAcked {
					log.Println("[micup-gateway] missed heartbeat ack; reconnecting")
					g.disconnectAndReconnect(false)
					return
				}
				g.sendHeartbeat()
			case <-g.heartbeatStop:
				return
			case <-g.ctx.Done():
				return
			}
		}
	}()
}

// stopHeartbeat stops the heartbeat ticker.
func (g *GatewayClient) stopHeartbeat() {
	if g.heartbeatStop != nil {
		close(g.heartbeatStop)
		g.heartbeatStop = nil
	}
}

// sendHeartbeat sends an OpHeartbeat packet with last sequence number.
func (g *GatewayClient) sendHeartbeat() {
	seq := atomic.LoadInt64(&g.lastSequence)
	var data []byte
	if seq > 0 {
		data, _ = json.Marshal(seq)
	} else {
		data = []byte("null")
	}

	g.heartbeatAcked = false
	g.lastHeartbeat = time.Now()

	_ = g.SendPayload(&GatewayPayload{
		Op: OpHeartbeat,
		D:  data,
	})
}

// SendPayload writes a raw payload onto the websocket connection with mutex synchronization.
func (g *GatewayClient) SendPayload(payload *GatewayPayload) error {
	g.writeMu.Lock()
	defer g.writeMu.Unlock()

	g.mu.Lock()
	conn := g.conn
	g.mu.Unlock()

	if conn == nil {
		return errors.New("gateway connection is closed")
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to serialize payload: %w", err)
	}

	return conn.WriteMessage(websocket.TextMessage, data)
}

// disconnectAndReconnect closes current connection and starts backoff reconnect.
func (g *GatewayClient) disconnectAndReconnect(clearSession bool) {
	g.mu.Lock()
	if clearSession {
		g.sessionID = ""
		g.lastSequence = 0
	}
	g.stopHeartbeat()
	if g.conn != nil {
		_ = g.conn.Close()
		g.conn = nil
	}
	atomic.StoreInt32(&g.status, statusDisconnected)
	g.mu.Unlock()

	go g.reconnectLoop()
}

// handleDisconnect is invoked when the websocket read loop terminates.
func (g *GatewayClient) handleDisconnect() {
	g.mu.Lock()
	g.stopHeartbeat()
	if g.conn != nil {
		_ = g.conn.Close()
		g.conn = nil
	}
	atomic.StoreInt32(&g.status, statusDisconnected)
	g.mu.Unlock()

	select {
	case <-g.ctx.Done():
		return
	default:
		go g.reconnectLoop()
	}
}

// reconnectLoop handles exponential backoff reconnects.
func (g *GatewayClient) reconnectLoop() {
	for {
		select {
		case <-g.ctx.Done():
			return
		default:
		}

		log.Printf("[micup-gateway] reconnecting in %v...", g.reconnectBackoff)
		time.Sleep(g.reconnectBackoff)

		g.reconnectBackoff *= 2
		if g.reconnectBackoff > 60*time.Second {
			g.reconnectBackoff = 60 * time.Second
		}

		if err := g.connect(); err == nil {
			log.Println("[micup-gateway] reconnected successfully!")
			return
		}
	}
}

// UpdateStatus sends OpPresenceUpdate to change bot activity and status.
func (g *GatewayClient) UpdateStatus(status string, activity *GatewayActivity) error {
	activities := []*GatewayActivity{}
	if activity != nil {
		activities = append(activities, activity)
	}

	presence := GatewayPresence{
		Status:     status,
		Activities: activities,
		AFK:        false,
	}

	data, err := json.Marshal(presence)
	if err != nil {
		return err
	}

	return g.SendPayload(&GatewayPayload{
		Op: OpPresenceUpdate,
		D:  data,
	})
}
