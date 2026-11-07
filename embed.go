package micup

import "time"

// EmbedBuilder provides a fluent, chained interface for constructing rich Embed cards.
type EmbedBuilder struct {
	embed *Embed
}

// NewEmbedBuilder creates an empty EmbedBuilder.
func NewEmbedBuilder() *EmbedBuilder {
	return &EmbedBuilder{
		embed: &Embed{
			Fields: make([]*EmbedField, 0),
		},
	}
}

// SetTitle sets the embed title.
func (b *EmbedBuilder) SetTitle(title string) *EmbedBuilder {
	b.embed.Title = title
	return b
}

// SetDescription sets the main text description.
func (b *EmbedBuilder) SetDescription(desc string) *EmbedBuilder {
	b.embed.Description = desc
	return b
}

// SetURL sets the URL the embed title links to.
func (b *EmbedBuilder) SetURL(url string) *EmbedBuilder {
	b.embed.URL = url
	return b
}

// SetColor sets the color stripe on the left of the embed (e.g. 0xFF5733).
func (b *EmbedBuilder) SetColor(color int) *EmbedBuilder {
	b.embed.Color = color
	return b
}

// SetTimestamp sets an ISO8601 timestamp string.
func (b *EmbedBuilder) SetTimestamp(t time.Time) *EmbedBuilder {
	b.embed.Timestamp = t.Format(time.RFC3339)
	return b
}

// SetFooter sets the footer text and optional icon URL.
func (b *EmbedBuilder) SetFooter(text string, iconURL string) *EmbedBuilder {
	b.embed.Footer = &EmbedFooter{
		Text:    text,
		IconURL: iconURL,
	}
	return b
}

// SetImage sets the main large image URL.
func (b *EmbedBuilder) SetImage(url string) *EmbedBuilder {
	b.embed.Image = &EmbedImage{URL: url}
	return b
}

// SetThumbnail sets the small thumbnail image URL on the top right.
func (b *EmbedBuilder) SetThumbnail(url string) *EmbedBuilder {
	b.embed.Thumbnail = &EmbedThumbnail{URL: url}
	return b
}

// SetAuthor sets the embed author name, URL, and avatar icon URL.
func (b *EmbedBuilder) SetAuthor(name, url, iconURL string) *EmbedBuilder {
	b.embed.Author = &EmbedAuthor{
		Name:    name,
		URL:     url,
		IconURL: iconURL,
	}
	return b
}

// AddField appends a titled field with inline flag.
func (b *EmbedBuilder) AddField(name, value string, inline bool) *EmbedBuilder {
	b.embed.Fields = append(b.embed.Fields, &EmbedField{
		Name:   name,
		Value:  value,
		Inline: inline,
	})
	return b
}

// Build returns the finalized Embed pointer.
func (b *EmbedBuilder) Build() *Embed {
	return b.embed
}
