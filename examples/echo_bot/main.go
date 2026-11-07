package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Qazcto7/micup-go"
)

func main() {
	token := os.Getenv("MICUP_BOT_TOKEN")
	if token == "" {
		log.Fatal("MICUP_BOT_TOKEN environment variable required")
	}

	sess, err := micup.New(token)
	if err != nil {
		log.Fatalf("Session creation failed: %v", err)
	}

	// Create command router with '!' prefix
	router := micup.NewCommandRouter("!")

	// Register 'echo' command
	router.Register(&micup.Command{
		Name:        "echo",
		Description: "Repeats text back to you.",
		Usage:       "<message>",
		Handler: func(ctx *micup.CommandContext) error {
			if len(ctx.Args) == 0 {
				_, err := ctx.Reply("⚠️ Please provide text to echo!")
				return err
			}
			text := strings.Join(ctx.Args, " ")
			_, err := ctx.Reply(fmt.Sprintf("🗣️ %s", text))
			return err
		},
	})

	// Register 'info' command with rich Embed
	router.Register(&micup.Command{
		Name:        "info",
		Description: "Shows bot information.",
		Handler: func(ctx *micup.CommandContext) error {
			embed := &micup.Embed{
				Title:       "Micup Go Bot",
				Description: "Powered by the official-grade micup-go SDK.",
				Color:       0x00FF88,
				Fields: []*micup.EmbedField{
					{
						Name:   "Author",
						Value:  ctx.Message.Author.Username,
						Inline: true,
					},
					{
						Name:   "Channel ID",
						Value:  ctx.Message.ChannelID,
						Inline: true,
					},
				},
				Footer: &micup.EmbedFooter{
					Text: "Built with morluto/rea reverse-engineering toolkit",
				},
			}
			_, err := ctx.ReplyEmbed(embed)
			return err
		},
	})

	// Attach router to session
	router.Attach(sess)

	// Connect to gateway
	if err := sess.Open(); err != nil {
		log.Fatalf("Failed to open connection: %v", err)
	}
	defer sess.Close()

	log.Println("Echo bot is online! Use !echo or !help")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Stopping...")
}
