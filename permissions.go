package micup

import (
	"context"
	"fmt"
	"strconv"
)

// Permission bit constants (64-bit bitmask).
const (
	PermissionCreateInstantInvite               int64 = 1 << 0
	PermissionKickMembers                       int64 = 1 << 1
	PermissionBanMembers                        int64 = 1 << 2
	PermissionAdministrator                     int64 = 1 << 3
	PermissionManageChannels                    int64 = 1 << 4
	PermissionManageGuild                       int64 = 1 << 5
	PermissionAddReactions                      int64 = 1 << 6
	PermissionViewAuditLog                      int64 = 1 << 7
	PermissionPrioritySpeaker                   int64 = 1 << 8
	PermissionStream                            int64 = 1 << 9
	PermissionViewChannel                       int64 = 1 << 10
	PermissionSendMessages                      int64 = 1 << 11
	PermissionSendTTSMessages                   int64 = 1 << 12
	PermissionManageMessages                    int64 = 1 << 13
	PermissionEmbedLinks                        int64 = 1 << 14
	PermissionAttachFiles                       int64 = 1 << 15
	PermissionReadMessageHistory                int64 = 1 << 16
	PermissionMentionEveryone                   int64 = 1 << 17
	PermissionUseExternalEmojis                 int64 = 1 << 18
	PermissionViewGuildInsights                 int64 = 1 << 19
	PermissionConnect                           int64 = 1 << 20
	PermissionSpeak                             int64 = 1 << 21
	PermissionMuteMembers                       int64 = 1 << 22
	PermissionDeafenMembers                     int64 = 1 << 23
	PermissionMoveMembers                       int64 = 1 << 24
	PermissionUseVAD                            int64 = 1 << 25
	PermissionChangeNickname                    int64 = 1 << 26
	PermissionManageNicknames                   int64 = 1 << 27
	PermissionManageRoles                       int64 = 1 << 28
	PermissionManageWebhooks                    int64 = 1 << 29
	PermissionManageGuildExpressions            int64 = 1 << 30
	PermissionUseApplicationCommands            int64 = 1 << 31
	PermissionRequestToSpeak                    int64 = 1 << 32
	PermissionManageEvents                      int64 = 1 << 33
	PermissionManageThreads                     int64 = 1 << 34
	PermissionCreatePublicThreads               int64 = 1 << 35
	PermissionCreatePrivateThreads              int64 = 1 << 36
	PermissionUseExternalStickers               int64 = 1 << 37
	PermissionSendMessagesInThreads             int64 = 1 << 38
	PermissionUseEmbeddedActivities             int64 = 1 << 39
	PermissionModerateMembers                   int64 = 1 << 40
	PermissionViewCreatorMonetizationAnalytics  int64 = 1 << 41
	PermissionUseSoundboard                     int64 = 1 << 42
	PermissionUseExternalSounds                 int64 = 1 << 45
	PermissionSendVoiceMessages                 int64 = 1 << 46
)

// ParsePermissions parses a string representation of permissions into int64.
func ParsePermissions(perms string) int64 {
	if perms == "" {
		return 0
	}
	val, err := strconv.ParseInt(perms, 10, 64)
	if err != nil {
		return 0
	}
	return val
}

// FormatPermissions formats int64 permissions into their string representation.
func FormatPermissions(perms int64) string {
	return strconv.FormatInt(perms, 10)
}

// HasPermission checks if the bitfield contains the specified permission or Administrator.
func HasPermission(bitfield int64, perm int64) bool {
	if (bitfield & PermissionAdministrator) == PermissionAdministrator {
		return true
	}
	return (bitfield & perm) == perm
}

// AddPermission adds permission flags to a bitfield.
func AddPermission(bitfield int64, perm int64) int64 {
	return bitfield | perm
}

// RemovePermission removes permission flags from a bitfield.
func RemovePermission(bitfield int64, perm int64) int64 {
	return bitfield &^ perm
}

// ComputeBasePermissions calculates a member's base permissions in a guild across all their roles.
func ComputeBasePermissions(member *Member, guild *Guild) int64 {
	if guild == nil || member == nil {
		return 0
	}

	// Guild owner has all permissions
	if guild.OwnerID == member.User.ID {
		return ^int64(0)
	}

	var permissions int64

	// 1. @everyone role (Role ID equals Guild ID)
	for _, role := range guild.Roles {
		if role.ID == guild.ID {
			permissions |= ParsePermissions(role.Permissions)
			break
		}
	}

	// 2. Member roles
	memberRoleMap := make(map[string]bool, len(member.Roles))
	for _, rID := range member.Roles {
		memberRoleMap[rID] = true
	}

	for _, role := range guild.Roles {
		if memberRoleMap[role.ID] {
			permissions |= ParsePermissions(role.Permissions)
		}
	}

	// If Administrator is present, bypass everything
	if (permissions & PermissionAdministrator) == PermissionAdministrator {
		return ^int64(0)
	}

	return permissions
}

// ComputeOverwrites applies channel-level permission overwrites to base permissions for a member.
func ComputeOverwrites(basePermissions int64, member *Member, channel *Channel) int64 {
	if channel == nil || member == nil {
		return basePermissions
	}

	// Administrator bypasses channel overwrites
	if (basePermissions & PermissionAdministrator) == PermissionAdministrator {
		return ^int64(0)
	}

	permissions := basePermissions

	// 1. Apply @everyone overwrite (ID == GuildID)
	for _, overwrite := range channel.PermissionOverwrites {
		if overwrite.ID == channel.GuildID {
			permissions &= ^ParsePermissions(overwrite.Deny)
			permissions |= ParsePermissions(overwrite.Allow)
			break
		}
	}

	// 2. Apply Role overwrites
	var roleAllow, roleDeny int64
	memberRoleMap := make(map[string]bool, len(member.Roles))
	for _, rID := range member.Roles {
		memberRoleMap[rID] = true
	}

	for _, overwrite := range channel.PermissionOverwrites {
		if overwrite.Type == 0 && memberRoleMap[overwrite.ID] { // 0 = Role
			roleDeny |= ParsePermissions(overwrite.Deny)
			roleAllow |= ParsePermissions(overwrite.Allow)
		}
	}

	permissions &= ^roleDeny
	permissions |= roleAllow

	// 3. Apply Member-specific overwrite (Type == 1)
	for _, overwrite := range channel.PermissionOverwrites {
		if overwrite.Type == 1 && overwrite.ID == member.User.ID { // 1 = Member
			permissions &= ^ParsePermissions(overwrite.Deny)
			permissions |= ParsePermissions(overwrite.Allow)
			break
		}
	}

	return permissions
}

// ComputeMemberPermissions calculates the effective permissions of a member in a specific channel.
func ComputeMemberPermissions(guild *Guild, channel *Channel, member *Member) int64 {
	base := ComputeBasePermissions(member, guild)
	return ComputeOverwrites(base, member, channel)
}

// ChannelPermissionEditParams specifies parameters for setting a channel permission overwrite.
type ChannelPermissionEditParams struct {
	Allow int64 `json:"allow,string"`
	Deny  int64 `json:"deny,string"`
	Type  int   `json:"type"` // 0 for role, 1 for member
}

// ChannelPermissionSet sets or updates a permission overwrite for a channel.
func (c *RESTClient) ChannelPermissionSet(ctx context.Context, channelID, overwriteID string, overwriteType int, allow, deny int64) error {
	params := ChannelPermissionEditParams{
		Allow: allow,
		Deny:  deny,
		Type:  overwriteType,
	}
	return c.RequestJSON(ctx, "PUT", fmt.Sprintf("/channels/%s/permissions/%s", channelID, overwriteID), params, nil)
}

// ChannelPermissionDelete deletes a permission overwrite for a channel.
func (c *RESTClient) ChannelPermissionDelete(ctx context.Context, channelID, overwriteID string) error {
	return c.RequestJSON(ctx, "DELETE", fmt.Sprintf("/channels/%s/permissions/%s", channelID, overwriteID), nil, nil)
}

// Session convenience wrappers

// ChannelPermissionSet sets channel permission overwrite.
func (s *Session) ChannelPermissionSet(channelID, overwriteID string, overwriteType int, allow, deny int64) error {
	return s.rest.ChannelPermissionSet(context.Background(), channelID, overwriteID, overwriteType, allow, deny)
}

// ChannelPermissionDelete removes a channel permission overwrite.
func (s *Session) ChannelPermissionDelete(channelID, overwriteID string) error {
	return s.rest.ChannelPermissionDelete(context.Background(), channelID, overwriteID)
}
