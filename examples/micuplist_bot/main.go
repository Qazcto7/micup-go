package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Qazcto7/micup-go"
)

func main() {
	micupToken := os.Getenv("MICUP_BOT_TOKEN")
	micupListToken := os.Getenv("MICUPLIST_API_TOKEN")
	webhookSecret := os.Getenv("WEBHOOK_SECRET")

	if micupToken == "" {
		log.Fatal("MICUP_BOT_TOKEN is required")
	}

	sess, err := micup.New(micupToken)
	if err != nil {
		log.Fatalf("Session error: %v", err)
	}

	listClient := micup.NewMicupListClient(micupListToken)

	// Post server count on READY
	sess.AddHandler(func(s *micup.Session, r *micup.Ready) {
		guildCount := len(r.Guilds)
		log.Printf("Bot online in %d guilds. Syncing with MicupList...", guildCount)

		if micupListToken != "" {
			ctx := context.Background()
			err := listClient.PostStats(ctx, r.User.ID, guildCount)
			if err != nil {
				log.Printf("Failed to post stats to MicupList: %v", err)
			} else {
				log.Println("Server count posted to MicupList successfully!")
			}
		}
	})

	// Vote webhook listener
	http.HandleFunc("/vote-webhook", micup.NewVoteWebhookHandler(webhookSecret, func(payload *micup.VoteWebhookPayload) {
		log.Printf("🎉 User %s voted for bot %s! (Weekend bonus: %v)",
			payload.UserID, payload.BotID, payload.IsWeekend)
	}))

	go func() {
		log.Println("Starting vote webhook server on :8080/vote-webhook")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Printf("Webhook server error: %v", err)
		}
	}()

	router := micup.NewCommandRouter("!")
	router.Register(&micup.Command{
		Name:        "vote",
		Description: "Check your vote status on MicupList.",
		Handler: func(ctx *micup.CommandContext) error {
			if micupListToken == "" {
				_, err := ctx.Reply("⚠️ MicupList API token is not configured on this bot.")
				return err
			}

			userMe, err := ctx.Session.UserMe()
			if err != nil {
				return err
			}

			res, err := listClient.CheckUserVote(context.Background(), userMe.ID, ctx.Message.Author.ID)
			if err != nil {
				return err
			}

			if res.Voted {
				_, err = ctx.Reply(fmt.Sprintf("✅ Thanks %s! You have voted for this bot.", ctx.Message.Author.Username))
			} else {
				_, err = ctx.Reply(fmt.Sprintf("❌ You haven't voted yet today, %s! Vote at https://micuplist.me", ctx.Message.Author.Username))
			}
			return err
		},
	})

	router.Attach(sess)

	if err := sess.Open(); err != nil {
		log.Fatalf("Gateway connect error: %v", err)
	}
	defer sess.Close()

	select {}
}
