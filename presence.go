package micup

// Presence status constants.
const (
	StatusOnline    = "online"
	StatusDND       = "dnd"
	StatusIdle      = "idle"
	StatusInvisible = "invisible"
	StatusOffline   = "offline"
)

// ActivityType constants.
const (
	ActivityTypeGame      = 0
	ActivityTypeStreaming = 1
	ActivityTypeListening = 2
	ActivityTypeWatching  = 3
	ActivityTypeCustom    = 4
	ActivityTypeCompeting = 5
)

// ActivityGame creates a game playing activity.
func ActivityGame(name string) *GatewayActivity {
	return &GatewayActivity{
		Name: name,
		Type: ActivityTypeGame,
	}
}

// ActivityStreaming creates a live streaming activity.
func ActivityStreaming(name, streamURL string) *GatewayActivity {
	return &GatewayActivity{
		Name: name,
		Type: ActivityTypeStreaming,
		URL:  streamURL,
	}
}

// ActivityListening creates a music/audio listening activity.
func ActivityListening(name string) *GatewayActivity {
	return &GatewayActivity{
		Name: name,
		Type: ActivityTypeListening,
	}
}

// ActivityWatching creates a video/movie watching activity.
func ActivityWatching(name string) *GatewayActivity {
	return &GatewayActivity{
		Name: name,
		Type: ActivityTypeWatching,
	}
}

// ActivityCustom creates a custom text status activity.
func ActivityCustom(text string) *GatewayActivity {
	return &GatewayActivity{
		Name:  "Custom Status",
		Type:  ActivityTypeCustom,
		State: text,
	}
}

// ActivityCompeting creates a competitive gaming activity.
func ActivityCompeting(name string) *GatewayActivity {
	return &GatewayActivity{
		Name: name,
		Type: ActivityTypeCompeting,
	}
}

// UpdateGameStatus updates the bot status and playing activity.
func (s *Session) UpdateGameStatus(status, game string) error {
	return s.gateway.UpdateStatus(status, ActivityGame(game))
}

// UpdateListeningStatus updates the bot status and listening activity.
func (s *Session) UpdateListeningStatus(status, track string) error {
	return s.gateway.UpdateStatus(status, ActivityListening(track))
}

// UpdateWatchingStatus updates the bot status and watching activity.
func (s *Session) UpdateWatchingStatus(status, show string) error {
	return s.gateway.UpdateStatus(status, ActivityWatching(show))
}

// UpdateCustomStatus updates the bot status with a custom text message.
func (s *Session) UpdateCustomStatus(status, text string) error {
	return s.gateway.UpdateStatus(status, ActivityCustom(text))
}

// UpdateStreamingStatus sets the bot status to streaming with title and Twitch/YouTube URL.
func (s *Session) UpdateStreamingStatus(title, streamURL string) error {
	return s.gateway.UpdateStatus(StatusOnline, ActivityStreaming(title, streamURL))
}
