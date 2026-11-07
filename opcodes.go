package micup

// Gateway OpCodes discovered from the Micup.gg (Fluxer) gateway protocol specification.
const (
	// OpDispatch (0) receives events from the gateway (READY, MESSAGE_CREATE, etc.)
	OpDispatch = 0

	// OpHeartbeat (1) is sent by the client to maintain connection, or requested by server.
	OpHeartbeat = 1

	// OpIdentify (2) starts a new gateway session during connection handshake.
	OpIdentify = 2

	// OpPresenceUpdate (3) updates the client's presence status and activity.
	OpPresenceUpdate = 3

	// OpVoiceStateUpdate (4) joins, leaves, or moves across voice channels.
	OpVoiceStateUpdate = 4

	// OpVoiceServerPing (5) keeps voice server connections alive.
	OpVoiceServerPing = 5

	// OpResume (6) resumes a previously disconnected session.
	OpResume = 6

	// OpReconnect (7) is sent by the server to instruct client to reconnect and resume.
	OpReconnect = 7

	// OpRequestGuildMembers (8) requests offline or all guild members.
	OpRequestGuildMembers = 8

	// OpInvalidSession (9) signals an invalid session; data payload indicates if resumable.
	OpInvalidSession = 9

	// OpHello (10) is received immediately upon connecting; contains heartbeat_interval.
	OpHello = 10

	// OpHeartbeatAck (11) is sent by server in response to a client OpHeartbeat.
	OpHeartbeatAck = 11

	// OpGatewayError (12) is sent by the server when a gateway error occurs.
	OpGatewayError = 12

	// OpLazyRequest (14) requests channel / viewport data lazily.
	OpLazyRequest = 14

	// OpRequestGuildCounts (15) requests guild member/online counts.
	OpRequestGuildCounts = 15

	// OpRequestChannelMemberCounts (16) requests member counts for channels.
	OpRequestChannelMemberCounts = 16

	// OpRequestForumUnreads (28) requests unread state for forum channels.
	OpRequestForumUnreads = 28
)

// Gateway Close Codes
const (
	CloseNormalClosure           = 1000
	CloseGoingAway               = 1001
	CloseProtocolError           = 1002
	CloseUnknownError            = 4000
	CloseUnknownOpcode           = 4001
	CloseDecodeError             = 4002
	CloseNotAuthenticated        = 4003
	CloseAuthenticationFailed   = 4004
	CloseAlreadyAuthenticated    = 4005
	CloseInvalidSeq              = 4007
	CloseRateLimited             = 4008
	CloseSessionTimedOut         = 4009
	CloseInvalidShard            = 4010
	CloseShardingRequired        = 4011
	CloseInvalidAPIVersion       = 4012
	CloseInvalidIntents          = 4013
	CloseDisallowedIntents       = 4014
)

// Gateway Dispatch Event Names
const (
	EventReady                 = "READY"
	EventResumed               = "RESUMED"
	EventMessageCreate         = "MESSAGE_CREATE"
	EventMessageUpdate         = "MESSAGE_UPDATE"
	EventMessageDelete         = "MESSAGE_DELETE"
	EventMessageBulkDelete     = "MESSAGE_BULK_DELETE"
	EventGuildCreate           = "GUILD_CREATE"
	EventGuildUpdate           = "GUILD_UPDATE"
	EventGuildDelete           = "GUILD_DELETE"
	EventGuildMemberAdd        = "GUILD_MEMBER_ADD"
	EventGuildMemberRemove     = "GUILD_MEMBER_REMOVE"
	EventGuildMemberUpdate     = "GUILD_MEMBER_UPDATE"
	EventGuildRoleCreate       = "GUILD_ROLE_CREATE"
	EventGuildRoleUpdate       = "GUILD_ROLE_UPDATE"
	EventGuildRoleDelete       = "GUILD_ROLE_DELETE"
	EventChannelCreate         = "CHANNEL_CREATE"
	EventChannelUpdate         = "CHANNEL_UPDATE"
	EventChannelDelete         = "CHANNEL_DELETE"
	EventChannelPinsUpdate     = "CHANNEL_PINS_UPDATE"
	EventTypingStart           = "TYPING_START"
	EventUserUpdate            = "USER_UPDATE"
	EventPresenceUpdate        = "PRESENCE_UPDATE"
	EventVoiceStateUpdate      = "VOICE_STATE_UPDATE"
	EventVoiceServerUpdate     = "VOICE_SERVER_UPDATE"
	EventInteractionCreate     = "INTERACTION_CREATE"
	EventInviteCreate          = "INVITE_CREATE"
	EventInviteDelete          = "INVITE_DELETE"
)
