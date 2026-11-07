package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Qazcto7/micup-go"
)

func main() {
	token := os.Getenv("MICUP_BOT_TOKEN")
	if token == "" {
		log.Fatal("Please set the MICUP_BOT_TOKEN environment variable")
	}

	// Initialize new session
	sess, err := micup.New(token)
	if err != nil {
		log.Fatalf("Error creating Micup session: %v", err)
	}

	// Register event handlers
	sess.AddHandler(func(s *micup.Session, r *micup.Ready) {
		log.Printf("Bot successfully logged in as: %s#%s (ID: %s)",
			r.User.Username, r.User.Discriminator, r.User.ID)

		// Set presence
		_ = s.UpdateStatus("online", &micup.GatewayActivity{
			Name: "!ping | micup-go",
			Type: 0,
		})
	})

	sess.AddHandler(func(s *micup.Session, m *micup.MessageCreate) {
		// Ignore bot's own messages
		if m.Author.Bot {
			return
		}

		if m.Content == "!ping" {
			reply := fmt.Sprintf("Pong! 🏓 Hello %s!", m.Author.Username)
			_, err := s.ChannelMessageSend(m.ChannelID, reply)
			if err != nil {
				log.Printf("Failed to send reply: %v", err)
			}
		}
	})

	// Open connection to gateway
	err = sess.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}
	defer sess.Close()

	log.Println("Bot is running. Press CTRL+C to exit.")

	// Wait for OS termination signal
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("Shutting down bot gracefully...")
}
