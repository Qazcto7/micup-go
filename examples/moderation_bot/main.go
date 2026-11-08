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
	token := os.Getenv("MICUP_TOKEN")
	if token == "" {
		log.Fatal("MICUP_TOKEN ortam değişkeni tanımlanmalı!")
	}

	session, err := micup.New(token)
	if err != nil {
		log.Fatalf("Oturum başlatılamadı: %v", err)
	}

	// 1. Bot-Side AutoMod Engine
	automod := micup.NewAutoModEngine()

	// Yasaklı kelime kuralı
	automod.AddKeywordRule("reklam_engeli", []string{"discord.gg/", "t.me/", "free-nitro"}, func(s *micup.Session, msg *micup.Message, reason string) {
		_ = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
		_, _ = s.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("⚠️ <@%s> Mesajınız silindi: %s", msg.Author.ID, reason))
	})

	// Aşırı etiketleme (Mass Mention) kuralı
	automod.AddMentionLimitRule("mass_mention", 4, func(s *micup.Session, msg *micup.Message, reason string) {
		_ = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
		_, _ = s.ChannelMessageSend(msg.ChannelID, fmt.Sprintf("🚨 <@%s> Çok fazla kullanıcı etiketlediniz (%s)!", msg.Author.ID, reason))
	})

	// AutoMod engine'i session'a bağla
	automod.Attach(session)

	// 2. Moderasyon Komutları
	session.AddHandler(func(s *micup.Session, e *micup.MessageCreate) {
		if e.Author == nil || e.Author.Bot {
			return
		}

		parts := strings.Fields(e.Content)
		if len(parts) == 0 {
			return
		}

		switch parts[0] {
		case "!audit":
			// Son denetim kayıtlarını listele
			if len(parts) < 2 {
				_, _ = s.ChannelMessageSend(e.ChannelID, "Kullanım: `!audit <guildID>`")
				return
			}
			guildID := parts[1]
			logs, err := s.GuildAuditLogs(guildID, &micup.AuditLogQuery{Limit: 5})
			if err != nil {
				_, _ = s.ChannelMessageSend(e.ChannelID, fmt.Sprintf("Denetim kaydı alınamadı: %v", err))
				return
			}

			embed := micup.NewEmbed().
				SetTitle("📋 Son 5 Denetim Kaydı").
				SetColor(0x3498db)

			for _, entry := range logs.Entries {
				embed.AddField(
					fmt.Sprintf("İşlem #%d (Hedef: %s)", entry.ActionType, entry.TargetID),
					fmt.Sprintf("Yapan: <@%s>\nSebep: %s", entry.UserID, entry.Reason),
					false,
				)
			}

			_, _ = s.ChannelMessageSendEmbed(e.ChannelID, embed.Build())

		case "!ban":
			if len(parts) < 3 {
				_, _ = s.ChannelMessageSend(e.ChannelID, "Kullanım: `!ban <guildID> <userID> [sebep]`")
				return
			}
			guildID := parts[1]
			targetID := parts[2]
			reason := "Kurallara uymamak"
			if len(parts) > 3 {
				reason = strings.Join(parts[3:], " ")
			}

			err := s.GuildMemberBan(guildID, targetID, 0, reason)
			if err != nil {
				_, _ = s.ChannelMessageSend(e.ChannelID, fmt.Sprintf("Yasaklama başarısız: %v", err))
				return
			}
			_, _ = s.ChannelMessageSend(e.ChannelID, fmt.Sprintf("🔨 <@%s> başarıyla yasaklandı. Sebep: %s", targetID, reason))
		}
	})

	if err := session.Open(); err != nil {
		log.Fatalf("Bağlantı hatası: %v", err)
	}
	defer session.Close()

	log.Println("Moderation Bot çalışıyor. Çıkmak için CTRL-C...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
