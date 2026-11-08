package micup

import (
	"errors"
	"sync"
)

var (
	ErrNilState      = errors.New("state is nil")
	ErrGuildNotFound = errors.New("guild not found in state cache")
	ErrChanNotFound  = errors.New("channel not found in state cache")
	ErrMemberNotFound = errors.New("member not found in state cache")
	ErrRoleNotFound   = errors.New("role not found in state cache")
	ErrMsgNotFound    = errors.New("message not found in state cache")
)

// State holds the in-memory cache of current entities (Guilds, Channels, Members, User, Messages).
type State struct {
	mu sync.RWMutex

	// The authenticated bot user
	User *User

	// Guild cache: guildID -> Guild
	Guilds map[string]*Guild

	// Channel cache: channelID -> Channel
	Channels map[string]*Channel

	// Member cache: guildID -> userID -> Member
	Members map[string]map[string]*Member

	// Role cache: guildID -> roleID -> Role
	Roles map[string]map[string]*Role

	// Recent message ring buffer / cache: channelID -> slice of messages
	Messages map[string][]*Message

	// MaxMessageHistory defines how many messages to retain per channel (default: 50)
	MaxMessageHistory int

	// Configuration flags
	TrackChannels bool
	TrackMembers  bool
	TrackRoles    bool
	TrackMessages bool
}

// NewState creates an initialized State cache instance.
func NewState() *State {
	return &State{
		Guilds:            make(map[string]*Guild),
		Channels:          make(map[string]*Channel),
		Members:           make(map[string]map[string]*Member),
		Roles:             make(map[string]map[string]*Role),
		Messages:          make(map[string][]*Message),
		MaxMessageHistory: 50,
		TrackChannels:     true,
		TrackMembers:      true,
		TrackRoles:        true,
		TrackMessages:     true,
	}
}

// BindSession registers state update listeners on the session's event bus.
func (s *State) BindSession(session *Session) {
	if s == nil || session == nil {
		return
	}

	session.AddHandler(func(r *Ready) {
		s.OnReady(r)
	})
	session.AddHandler(func(g *GuildCreate) {
		s.OnGuildCreate(g)
	})
	session.AddHandler(func(g *GuildUpdate) {
		s.OnGuildUpdate(g)
	})
	session.AddHandler(func(g *GuildDelete) {
		s.OnGuildDelete(g)
	})
	session.AddHandler(func(c *ChannelCreate) {
		s.OnChannelCreate(c)
	})
	session.AddHandler(func(c *ChannelUpdate) {
		s.OnChannelUpdate(c)
	})
	session.AddHandler(func(c *ChannelDelete) {
		s.OnChannelDelete(c)
	})
	session.AddHandler(func(m *GuildMemberAdd) {
		s.OnGuildMemberAdd(m)
	})
	session.AddHandler(func(m *GuildMemberRemove) {
		s.OnGuildMemberRemove(m)
	})
	session.AddHandler(func(m *GuildMemberUpdate) {
		s.OnGuildMemberUpdate(m)
	})
	session.AddHandler(func(m *MessageCreate) {
		s.OnMessageCreate(m)
	})
	session.AddHandler(func(m *MessageDelete) {
		s.OnMessageDelete(m)
	})
	session.AddHandler(func(m *MessageBulkDelete) {
		s.OnMessageBulkDelete(m)
	})
}

// GetUser returns the cached bot user.
func (s *State) GetUser() *User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.User
}

// Guild retrieves a guild by ID from cache.
func (s *State) Guild(guildID string) (*Guild, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	g, exists := s.Guilds[guildID]
	if !exists {
		return nil, ErrGuildNotFound
	}
	return g, nil
}

// GuildsList returns a copy slice of all cached guilds.
func (s *State) GuildsList() []*Guild {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*Guild, 0, len(s.Guilds))
	for _, g := range s.Guilds {
		list = append(list, g)
	}
	return list
}

// Channel retrieves a channel by ID from cache.
func (s *State) Channel(channelID string) (*Channel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, exists := s.Channels[channelID]
	if !exists {
		return nil, ErrChanNotFound
	}
	return c, nil
}

// Member retrieves a guild member from cache.
func (s *State) Member(guildID, userID string) (*Member, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	guildMembers, exists := s.Members[guildID]
	if !exists {
		return nil, ErrMemberNotFound
	}
	member, exists := guildMembers[userID]
	if !exists {
		return nil, ErrMemberNotFound
	}
	return member, nil
}

// MembersList returns all cached members for a guild.
func (s *State) MembersList(guildID string) []*Member {
	s.mu.RLock()
	defer s.mu.RUnlock()

	guildMembers, exists := s.Members[guildID]
	if !exists {
		return nil
	}

	list := make([]*Member, 0, len(guildMembers))
	for _, m := range guildMembers {
		list = append(list, m)
	}
	return list
}

// Role retrieves a role from cache.
func (s *State) Role(guildID, roleID string) (*Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	guildRoles, exists := s.Roles[guildID]
	if !exists {
		return nil, ErrRoleNotFound
	}
	role, exists := guildRoles[roleID]
	if !exists {
		return nil, ErrRoleNotFound
	}
	return role, nil
}

// Message retrieves a cached message from a channel.
func (s *State) Message(channelID, messageID string) (*Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	msgs, exists := s.Messages[channelID]
	if !exists {
		return nil, ErrMsgNotFound
	}
	for _, m := range msgs {
		if m.ID == messageID {
			return m, nil
		}
	}
	return nil, ErrMsgNotFound
}

// MessagesList returns all cached messages for a channel.
func (s *State) MessagesList(channelID string) []*Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	msgs, exists := s.Messages[channelID]
	if !exists {
		return nil
	}

	res := make([]*Message, len(msgs))
	copy(res, msgs)
	return res
}

// Mutator event handlers

func (s *State) OnReady(r *Ready) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if r != nil && r.User != nil {
		s.User = r.User
	}
}

func (s *State) OnGuildCreate(g *GuildCreate) {
	if g == nil || g.Guild == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.Guilds[g.ID] = g.Guild

	if s.TrackChannels && len(g.Channels) > 0 {
		for _, ch := range g.Channels {
			ch.GuildID = g.ID
			s.Channels[ch.ID] = ch
		}
	}

	if s.TrackRoles && len(g.Roles) > 0 {
		if _, exists := s.Roles[g.ID]; !exists {
			s.Roles[g.ID] = make(map[string]*Role)
		}
		for _, r := range g.Roles {
			s.Roles[g.ID][r.ID] = r
		}
	}

	if s.TrackMembers && len(g.Members) > 0 {
		if _, exists := s.Members[g.ID]; !exists {
			s.Members[g.ID] = make(map[string]*Member)
		}
		for _, m := range g.Members {
			if m.User != nil {
				m.GuildID = g.ID
				s.Members[g.ID][m.User.ID] = m
			}
		}
	}
}

func (s *State) OnGuildUpdate(g *GuildUpdate) {
	if g == nil || g.Guild == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.Guilds[g.ID]
	if exists {
		// Keep cached channels/roles if not present in partial update
		if len(g.Channels) == 0 {
			g.Channels = existing.Channels
		}
		if len(g.Roles) == 0 {
			g.Roles = existing.Roles
		}
		if len(g.Members) == 0 {
			g.Members = existing.Members
		}
	}
	s.Guilds[g.ID] = g.Guild
}

func (s *State) OnGuildDelete(g *GuildDelete) {
	if g == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.Guilds, g.ID)
	delete(s.Members, g.ID)
	delete(s.Roles, g.ID)

	// Clean associated channels
	for chID, ch := range s.Channels {
		if ch.GuildID == g.ID {
			delete(s.Channels, chID)
			delete(s.Messages, chID)
		}
	}
}

func (s *State) OnChannelCreate(c *ChannelCreate) {
	if c == nil || c.Channel == nil || !s.TrackChannels {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.Channels[c.ID] = c.Channel

	// Append to guild channels if guild exists
	if c.GuildID != "" {
		if g, exists := s.Guilds[c.GuildID]; exists {
			g.Channels = append(g.Channels, c.Channel)
		}
	}
}

func (s *State) OnChannelUpdate(c *ChannelUpdate) {
	if c == nil || c.Channel == nil || !s.TrackChannels {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.Channels[c.ID] = c.Channel
}

func (s *State) OnChannelDelete(c *ChannelDelete) {
	if c == nil || c.Channel == nil || !s.TrackChannels {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.Channels, c.ID)
	delete(s.Messages, c.ID)

	if c.GuildID != "" {
		if g, exists := s.Guilds[c.GuildID]; exists {
			for i, ch := range g.Channels {
				if ch.ID == c.ID {
					g.Channels = append(g.Channels[:i], g.Channels[i+1:]...)
					break
				}
			}
		}
	}
}

func (s *State) OnGuildMemberAdd(m *GuildMemberAdd) {
	if m == nil || m.Member == nil || m.User == nil || !s.TrackMembers {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Members[m.GuildID]; !exists {
		s.Members[m.GuildID] = make(map[string]*Member)
	}
	m.Member.GuildID = m.GuildID
	s.Members[m.GuildID][m.User.ID] = m.Member
}

func (s *State) OnGuildMemberRemove(m *GuildMemberRemove) {
	if m == nil || m.User == nil || !s.TrackMembers {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if guildMembers, exists := s.Members[m.GuildID]; exists {
		delete(guildMembers, m.User.ID)
	}
}

func (s *State) OnGuildMemberUpdate(m *GuildMemberUpdate) {
	if m == nil || m.User == nil || !s.TrackMembers {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if guildMembers, exists := s.Members[m.GuildID]; exists {
		if existing, exists := guildMembers[m.User.ID]; exists {
			existing.Roles = m.Roles
			if m.Nick != "" {
				existing.Nick = m.Nick
			}
		}
	}
}

func (s *State) OnMessageCreate(m *MessageCreate) {
	if m == nil || m.Message == nil || !s.TrackMessages || s.MaxMessageHistory <= 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	list := s.Messages[m.ChannelID]
	list = append(list, m.Message)

	// Trim older messages if above max
	if len(list) > s.MaxMessageHistory {
		list = list[len(list)-s.MaxMessageHistory:]
	}
	s.Messages[m.ChannelID] = list
}

func (s *State) OnMessageDelete(m *MessageDelete) {
	if m == nil || !s.TrackMessages {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	list, exists := s.Messages[m.ChannelID]
	if !exists {
		return
	}

	for i, msg := range list {
		if msg.ID == m.ID {
			s.Messages[m.ChannelID] = append(list[:i], list[i+1:]...)
			break
		}
	}
}

func (s *State) OnMessageBulkDelete(m *MessageBulkDelete) {
	if m == nil || len(m.IDs) == 0 || !s.TrackMessages {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	list, exists := s.Messages[m.ChannelID]
	if !exists {
		return
	}

	idMap := make(map[string]struct{}, len(m.IDs))
	for _, id := range m.IDs {
		idMap[id] = struct{}{}
	}

	filtered := make([]*Message, 0, len(list))
	for _, msg := range list {
		if _, deleted := idMap[msg.ID]; !deleted {
			filtered = append(filtered, msg)
		}
	}
	s.Messages[m.ChannelID] = filtered
}
