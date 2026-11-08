package micup

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// Scheduled Event Status constants.
const (
	EventStatusScheduled = 1
	EventStatusActive    = 2
	EventStatusCompleted = 3
	EventStatusCanceled  = 4
)

// Scheduled Event Entity Type constants.
const (
	EventEntityStageInstance = 1
	EventEntityVoice         = 2
	EventEntityExternal      = 3
)

// Scheduled Event Privacy Level constants.
const (
	EventPrivacyGuildOnly = 2
)

// ScheduledEventEntityMetadata represents additional metadata for external events.
type ScheduledEventEntityMetadata struct {
	Location string `json:"location,omitempty"`
}

// GuildScheduledEvent represents a planned event in a guild.
type GuildScheduledEvent struct {
	ID                 string                        `json:"id"`
	GuildID            string                        `json:"guild_id"`
	ChannelID          string                        `json:"channel_id,omitempty"`
	CreatorID          string                        `json:"creator_id,omitempty"`
	Name               string                        `json:"name"`
	Description        string                        `json:"description,omitempty"`
	ScheduledStartTime time.Time                     `json:"scheduled_start_time"`
	ScheduledEndTime   *time.Time                    `json:"scheduled_end_time,omitempty"`
	PrivacyLevel       int                           `json:"privacy_level"`
	Status             int                           `json:"status"`
	EntityType         int                           `json:"entity_type"`
	EntityID           string                        `json:"entity_id,omitempty"`
	EntityMetadata     *ScheduledEventEntityMetadata `json:"entity_metadata,omitempty"`
	Creator            *User                         `json:"creator,omitempty"`
	UserCount          int                           `json:"user_count,omitempty"`
	Image              string                        `json:"image,omitempty"`
}

// CreateScheduledEventParams defines parameters for creating a scheduled event.
type CreateScheduledEventParams struct {
	ChannelID          string                        `json:"channel_id,omitempty"`
	EntityMetadata     *ScheduledEventEntityMetadata `json:"entity_metadata,omitempty"`
	Name               string                        `json:"name"`
	PrivacyLevel       int                           `json:"privacy_level"`
	ScheduledStartTime time.Time                     `json:"scheduled_start_time"`
	ScheduledEndTime   *time.Time                    `json:"scheduled_end_time,omitempty"`
	Description        string                        `json:"description,omitempty"`
	EntityType         int                           `json:"entity_type"`
	Image              string                        `json:"image,omitempty"`
}

// ModifyScheduledEventParams defines parameters for updating an existing scheduled event.
type ModifyScheduledEventParams struct {
	ChannelID          *string                       `json:"channel_id,omitempty"`
	EntityMetadata     *ScheduledEventEntityMetadata `json:"entity_metadata,omitempty"`
	Name               *string                       `json:"name,omitempty"`
	PrivacyLevel       *int                          `json:"privacy_level,omitempty"`
	ScheduledStartTime *time.Time                    `json:"scheduled_start_time,omitempty"`
	ScheduledEndTime   *time.Time                    `json:"scheduled_end_time,omitempty"`
	Description        *string                       `json:"description,omitempty"`
	EntityType         *int                          `json:"entity_type,omitempty"`
	Status             *int                          `json:"status,omitempty"`
	Image              *string                       `json:"image,omitempty"`
}

// GuildScheduledEventUser represents a user registered for a scheduled event.
type GuildScheduledEventUser struct {
	GuildScheduledEventID string  `json:"guild_scheduled_event_id"`
	User                  *User   `json:"user"`
	Member                *Member `json:"member,omitempty"`
}

// GuildScheduledEvents fetches all scheduled events for a guild.
func (c *RESTClient) GuildScheduledEvents(ctx context.Context, guildID string, withUserCount bool) ([]*GuildScheduledEvent, error) {
	endpoint := fmt.Sprintf("/guilds/%s/scheduled-events", guildID)
	if withUserCount {
		endpoint += "?with_user_count=true"
	}

	var events []*GuildScheduledEvent
	err := c.RequestJSON(ctx, "GET", endpoint, nil, &events)
	if err != nil {
		return nil, err
	}
	return events, nil
}

// GuildScheduledEvent fetches a single scheduled event by ID.
func (c *RESTClient) GuildScheduledEvent(ctx context.Context, guildID, eventID string, withUserCount bool) (*GuildScheduledEvent, error) {
	endpoint := fmt.Sprintf("/guilds/%s/scheduled-events/%s", guildID, eventID)
	if withUserCount {
		endpoint += "?with_user_count=true"
	}

	var event GuildScheduledEvent
	err := c.RequestJSON(ctx, "GET", endpoint, nil, &event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// GuildScheduledEventCreate creates a new scheduled event.
func (c *RESTClient) GuildScheduledEventCreate(ctx context.Context, guildID string, params *CreateScheduledEventParams) (*GuildScheduledEvent, error) {
	var event GuildScheduledEvent
	err := c.RequestJSON(ctx, "POST", fmt.Sprintf("/guilds/%s/scheduled-events", guildID), params, &event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// GuildScheduledEventModify updates an existing scheduled event.
func (c *RESTClient) GuildScheduledEventModify(ctx context.Context, guildID, eventID string, params *ModifyScheduledEventParams) (*GuildScheduledEvent, error) {
	var event GuildScheduledEvent
	err := c.RequestJSON(ctx, "PATCH", fmt.Sprintf("/guilds/%s/scheduled-events/%s", guildID, eventID), params, &event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// GuildScheduledEventDelete deletes a scheduled event.
func (c *RESTClient) GuildScheduledEventDelete(ctx context.Context, guildID, eventID string) error {
	return c.RequestJSON(ctx, "DELETE", fmt.Sprintf("/guilds/%s/scheduled-events/%s", guildID, eventID), nil, nil)
}

// GuildScheduledEventUsers lists users subscribed to a scheduled event.
func (c *RESTClient) GuildScheduledEventUsers(ctx context.Context, guildID, eventID string, limit int, withMember bool, before, after string) ([]*GuildScheduledEventUser, error) {
	endpoint := fmt.Sprintf("/guilds/%s/scheduled-events/%s/users", guildID, eventID)
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if withMember {
		params.Set("with_member", "true")
	}
	if before != "" {
		params.Set("before", before)
	}
	if after != "" {
		params.Set("after", after)
	}
	if len(params) > 0 {
		endpoint = fmt.Sprintf("%s?%s", endpoint, params.Encode())
	}

	var users []*GuildScheduledEventUser
	err := c.RequestJSON(ctx, "GET", endpoint, nil, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// Session convenience methods

// GuildScheduledEvents lists scheduled events in a guild.
func (s *Session) GuildScheduledEvents(guildID string, withUserCount bool) ([]*GuildScheduledEvent, error) {
	return s.rest.GuildScheduledEvents(context.Background(), guildID, withUserCount)
}

// GuildScheduledEvent fetches a single scheduled event.
func (s *Session) GuildScheduledEvent(guildID, eventID string, withUserCount bool) (*GuildScheduledEvent, error) {
	return s.rest.GuildScheduledEvent(context.Background(), guildID, eventID, withUserCount)
}

// GuildScheduledEventCreate creates a scheduled event.
func (s *Session) GuildScheduledEventCreate(guildID string, params *CreateScheduledEventParams) (*GuildScheduledEvent, error) {
	return s.rest.GuildScheduledEventCreate(context.Background(), guildID, params)
}

// GuildScheduledEventModify modifies a scheduled event.
func (s *Session) GuildScheduledEventModify(guildID, eventID string, params *ModifyScheduledEventParams) (*GuildScheduledEvent, error) {
	return s.rest.GuildScheduledEventModify(context.Background(), guildID, eventID, params)
}

// GuildScheduledEventDelete cancels and deletes a scheduled event.
func (s *Session) GuildScheduledEventDelete(guildID, eventID string) error {
	return s.rest.GuildScheduledEventDelete(context.Background(), guildID, eventID)
}

// GuildScheduledEventUsers returns users interested in a scheduled event.
func (s *Session) GuildScheduledEventUsers(guildID, eventID string, limit int) ([]*GuildScheduledEventUser, error) {
	return s.rest.GuildScheduledEventUsers(context.Background(), guildID, eventID, limit, true, "", "")
}
