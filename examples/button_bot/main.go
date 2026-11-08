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
		log.Fatal("MICUP_BOT_TOKEN environment variable not set")
	}

	client, err := micup.New(token)
	if err != nil {
		log.Fatalf("Failed to initialize bot: %v", err)
	}

	// Set dynamic bot presence
	client.AddHandler(func(s *micup.Session, r *micup.Ready) {
		log.Printf("Bot ready as %s#%s (ID: %s)", r.User.Username, r.User.Discriminator, r.User.ID)
		s.UpdateGameStatus(micup.StatusOnline, "Micup.gg | !panel")
	})

	// Setup component router for interactive buttons
	compRouter := micup.NewComponentRouter()

	compRouter.Handle("btn_accept", func(s *micup.Session, i *micup.Interaction) {
		s.InteractionRespond(i, &micup.InteractionResponse{
			Type: micup.InteractionResponseChannelMessageWithSource,
			Data: &micup.InteractionResponseData{
				Content: "✅ Şartları kabul ettiniz! Rolünüz tanımlandı.",
			},
		})
	})

	compRouter.Handle("btn_reject", func(s *micup.Session, i *micup.Interaction) {
		s.InteractionRespond(i, &micup.InteractionResponse{
			Type: micup.InteractionResponseChannelMessageWithSource,
			Data: &micup.InteractionResponseData{
				Content: "❌ Şartları reddettiniz.",
			},
		})
	})

	// Register component router listener
	client.AddHandler(compRouter.Listener())

	// Listen for !panel command to post interactive button card
	client.AddHandler(func(s *micup.Session, m *micup.MessageCreate) {
		if m.Author.Bot || m.Content != "!panel" {
			return
		}

		embed := micup.NewEmbed().
			SetTitle("📋 Topluluk Kuralları & Onay Paneli").
			SetDescription("Sunucumuza hoş geldiniz! Kanallara erişmek için lütfen kuralları onaylayın.").
			SetColor(0x5865F2).
			AddField("1. Saygı", "Tüm üyelere karşı nezaketli olun.", false).
			AddField("2. Spam Yok", "Kanalları flood ve spam ile meşgul etmeyin.", false).
			SetFooter("Micup Bot Güvenlik Sistemi", "").
			Build()

		actionRow := micup.NewActionRow(
			micup.NewSuccessButton("btn_accept", "Kabul Et").WithEmoji("✅"),
			micup.NewDangerButton("btn_reject", "Reddet").WithEmoji("❌"),
			micup.NewLinkButton("https://micup.gg", "Web Sitesi"),
		)

		msgSend := &micup.MessageSend{
			Embeds:     []*micup.Embed{embed},
			Components: []micup.Component{actionRow},
		}

		_, err := s.ChannelMessageSendComplex(m.ChannelID, msgSend)
		if err != nil {
			log.Printf("Failed to send panel: %v", err)
		}
	})

	if err := client.Open(); err != nil {
		log.Fatalf("Failed to connect to gateway: %v", err)
	}
	defer client.Close()

	fmt.Println("Button bot is now running. Press Ctrl+C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
