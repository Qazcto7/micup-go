package micup

import (
	"context"
)

// MessageIterator automatically paginates through channel message history.
type MessageIterator struct {
	client    *RESTClient
	channelID string
	limit     int
	before    string
	after     string
	buffer    []*Message
	done      bool
}

// NewMessageIterator creates an iterator over a channel's messages.
func (c *RESTClient) NewMessageIterator(channelID string, chunkSize int, before, after string) *MessageIterator {
	if chunkSize <= 0 || chunkSize > 100 {
		chunkSize = 100
	}
	return &MessageIterator{
		client:    c,
		channelID: channelID,
		limit:     chunkSize,
		before:    before,
		after:     after,
		buffer:    make([]*Message, 0),
		done:      false,
	}
}

// Next returns the next message in history, fetching new pages as needed.
func (it *MessageIterator) Next(ctx context.Context) (*Message, error) {
	if len(it.buffer) > 0 {
		msg := it.buffer[0]
		it.buffer = it.buffer[1:]
		return msg, nil
	}

	if it.done {
		return nil, nil
	}

	// Fetch next page
	msgs, err := it.client.ChannelMessages(ctx, it.channelID, it.limit, "", it.before, it.after)
	if err != nil {
		return nil, err
	}

	if len(msgs) == 0 {
		it.done = true
		return nil, nil
	}

	if len(msgs) < it.limit {
		it.done = true
	}

	// Update cursor for backward pagination
	if it.after == "" {
		it.before = msgs[len(msgs)-1].ID
	} else {
		it.after = msgs[0].ID
	}

	it.buffer = msgs[1:]
	return msgs[0], nil
}

// All fetches all remaining messages in the iterator up to maxCount.
func (it *MessageIterator) All(ctx context.Context, maxCount int) ([]*Message, error) {
	var results []*Message
	for {
		if maxCount > 0 && len(results) >= maxCount {
			break
		}

		msg, err := it.Next(ctx)
		if err != nil {
			return results, err
		}
		if msg == nil {
			break
		}
		results = append(results, msg)
	}
	return results, nil
}

// MemberIterator automatically paginates through all members of a guild.
type MemberIterator struct {
	client  *RESTClient
	guildID string
	limit   int
	after   string
	buffer  []*Member
	done    bool
}

// NewMemberIterator creates an iterator over a guild's members.
func (c *RESTClient) NewMemberIterator(guildID string, chunkSize int) *MemberIterator {
	if chunkSize <= 0 || chunkSize > 1000 {
		chunkSize = 1000
	}
	return &MemberIterator{
		client:  c,
		guildID: guildID,
		limit:   chunkSize,
		buffer:  make([]*Member, 0),
		done:    false,
	}
}

// Next returns the next member, fetching pages as needed.
func (it *MemberIterator) Next(ctx context.Context) (*Member, error) {
	if len(it.buffer) > 0 {
		m := it.buffer[0]
		it.buffer = it.buffer[1:]
		return m, nil
	}

	if it.done {
		return nil, nil
	}

	members, err := it.client.GetGuildMembers(ctx, it.guildID, it.limit, it.after)
	if err != nil {
		return nil, err
	}

	if len(members) == 0 {
		it.done = true
		return nil, nil
	}

	if len(members) < it.limit {
		it.done = true
	}

	lastMember := members[len(members)-1]
	if lastMember.User != nil {
		it.after = lastMember.User.ID
	}

	it.buffer = members[1:]
	return members[0], nil
}

// All fetches all members in the guild.
func (it *MemberIterator) All(ctx context.Context, maxCount int) ([]*Member, error) {
	var results []*Member
	for {
		if maxCount > 0 && len(results) >= maxCount {
			break
		}

		m, err := it.Next(ctx)
		if err != nil {
			return results, err
		}
		if m == nil {
			break
		}
		results = append(results, m)
	}
	return results, nil
}
