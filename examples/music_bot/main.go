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

	var activePlayer *micup.AudioPlayer

	session.AddHandler(func(s *micup.Session, e *micup.MessageCreate) {
		if e.Author == nil || e.Author.Bot {
			return
		}

		parts := strings.Fields(e.Content)
		if len(parts) == 0 {
			return
		}

		cmd := parts[0]

		switch cmd {
		case "!join":
			if len(parts) < 3 {
				_, _ = s.ChannelMessageSend(e.ChannelID, "Kullanım: `!join <guildID> <channelID>`")
				return
			}
			guildID := parts[1]
			channelID := parts[2]

			vc, err := s.VoiceJoin(guildID, channelID, false, false)
			if err != nil {
				_, _ = s.ChannelMessageSend(e.ChannelID, fmt.Sprintf("Ses kanalına bağlanılamadı: %v", err))
				return
			}

			activePlayer = vc.NewPlayer()
			activePlayer.OnFinish(func() {
				_, _ = s.ChannelMessageSend(e.ChannelID, "🎵 Çalma listesi sona erdi.")
			})

			_, _ = s.ChannelMessageSend(e.ChannelID, fmt.Sprintf("✅ Ses kanalına bağlandı! (Oda: %s)", vc.RoomName))

		case "!play":
			if activePlayer == nil {
				_, _ = s.ChannelMessageSend(e.ChannelID, "Önce bir ses kanalına girmeliyim: `!join <guildID> <channelID>`")
				return
			}
			if len(parts) < 2 {
				_, _ = s.ChannelMessageSend(e.ChannelID, "Kullanım: `!play <dosya_veya_link>`")
				return
			}
			source := parts[1]
			if err := activePlayer.Play(source); err != nil {
				_, _ = s.ChannelMessageSend(e.ChannelID, fmt.Sprintf("Oynatma hatası: %v", err))
				return
			}
			_, _ = s.ChannelMessageSend(e.ChannelID, fmt.Sprintf("▶️ Oynatılıyor: %s", source))

		case "!pause":
			if activePlayer != nil {
				activePlayer.Pause()
				_, _ = s.ChannelMessageSend(e.ChannelID, "⏸️ Oynatma duraklatıldı.")
			}

		case "!resume":
			if activePlayer != nil {
				activePlayer.Resume()
				_, _ = s.ChannelMessageSend(e.ChannelID, "▶️ Oynatma devam ettiriliyor.")
			}

		case "!stop":
			if activePlayer != nil {
				activePlayer.Stop()
				_, _ = s.ChannelMessageSend(e.ChannelID, "⏹️ Oynatma durduruldu.")
			}

		case "!volume":
			if activePlayer == nil {
				return
			}
			if len(parts) < 2 {
				_, _ = s.ChannelMessageSend(e.ChannelID, fmt.Sprintf("Mevcut ses düzeyi: %%%.0f", activePlayer.Volume()*100))
				return
			}
			var vol float64
			fmt.Sscanf(parts[1], "%f", &vol)
			activePlayer.SetVolume(vol / 100.0)
			_, _ = s.ChannelMessageSend(e.ChannelID, fmt.Sprintf("🔊 Ses düzeyi ayarlandı: %%%.0f", vol))
		}
	})

	if err := session.Open(); err != nil {
		log.Fatalf("Bağlantı hatası: %v", err)
	}
	defer session.Close()

	log.Println("Music Bot çalışıyor. Çıkmak için CTRL-C...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
