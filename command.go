package micup

import (
	"fmt"
	"strings"
	"sync"
)

// CommandContext represents the invocation context of a message command.
type CommandContext struct {
	Session *Session
	Message *Message
	Args    []string
	Command string
}

// Reply sends a reply message to the current channel.
func (c *CommandContext) Reply(content string) (*Message, error) {
	return c.Session.ChannelMessageSend(c.Message.ChannelID, content)
}

// ReplyEmbed sends an embed reply to the current channel.
func (c *CommandContext) ReplyEmbed(embed *Embed) (*Message, error) {
	return c.Session.ChannelMessageSendEmbed(c.Message.ChannelID, embed)
}

// CommandHandlerFunc is the callback signature for a command.
type CommandHandlerFunc func(ctx *CommandContext) error

// Command holds registered command details.
type Command struct {
	Name        string
	Description string
	Usage       string
	Aliases     []string
	Handler     CommandHandlerFunc
}

// CommandRouter routes message create events to matching commands.
type CommandRouter struct {
	Prefix   string
	commands map[string]*Command
	mu       sync.RWMutex
}

// NewCommandRouter initializes a command router with a command prefix (e.g. "!").
func NewCommandRouter(prefix string) *CommandRouter {
	r := &CommandRouter{
		Prefix:   prefix,
		commands: make(map[string]*Command),
	}

	// Register built-in help command
	r.Register(&Command{
		Name:        "help",
		Description: "Displays list of available commands and usage instructions.",
		Handler: func(ctx *CommandContext) error {
			lines := []string{"**Available Commands:**"}
			r.mu.RLock()
			for _, cmd := range r.commands {
				if cmd.Name == "help" {
					continue
				}
				line := fmt.Sprintf("• `%s%s` - %s", r.Prefix, cmd.Name, cmd.Description)
				if cmd.Usage != "" {
					line += fmt.Sprintf(" (Usage: `%s%s %s`)", r.Prefix, cmd.Name, cmd.Usage)
				}
				lines = append(lines, line)
			}
			r.mu.RUnlock()

			_, err := ctx.Reply(strings.Join(lines, "\n"))
			return err
		},
	})

	return r
}

// Register registers a command into the router.
func (r *CommandRouter) Register(cmd *Command) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.commands[strings.ToLower(cmd.Name)] = cmd
	for _, alias := range cmd.Aliases {
		r.commands[strings.ToLower(alias)] = cmd
	}
}

// Attach attaches this router as an event handler to a Micup session.
func (r *CommandRouter) Attach(s *Session) {
	s.AddHandler(func(sess *Session, m *MessageCreate) {
		r.HandleMessage(sess, m.Message)
	})
}

// HandleMessage evaluates a message and executes matching commands.
func (r *CommandRouter) HandleMessage(s *Session, msg *Message) {
	if msg == nil || msg.Author == nil || msg.Author.Bot {
		return // Ignore bot self-messages and empty authors
	}

	content := strings.TrimSpace(msg.Content)
	if !strings.HasPrefix(content, r.Prefix) {
		return
	}

	rawCmd := strings.TrimPrefix(content, r.Prefix)
	parts := strings.Fields(rawCmd)
	if len(parts) == 0 {
		return
	}

	cmdName := strings.ToLower(parts[0])
	args := parts[1:]

	r.mu.RLock()
	cmd, exists := r.commands[cmdName]
	r.mu.RUnlock()

	if !exists {
		return
	}

	ctx := &CommandContext{
		Session: s,
		Message: msg,
		Args:    args,
		Command: cmdName,
	}

	if err := cmd.Handler(ctx); err != nil {
		_, _ = ctx.Reply(fmt.Sprintf("❌ Error executing `%s`: %v", cmdName, err))
	}
}
