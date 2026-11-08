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
	token := os.Getenv("MICUP_TOKEN")
	if token == "" {
		log.Fatal("MICUP_TOKEN ortam değişkeni tanımlanmalı!")
	}

	session, err := micup.New(token)
	if err != nil {
		log.Fatalf("Oturum başlatılamadı: %v", err)
	}

	// 1. Destek / Ticket Yöneticisi Konfigürasyonu
	ticketMgr := micup.NewTicketManager(session, &micup.TicketConfig{
		GuildID:           os.Getenv("MICUP_GUILD_ID"), // Ticket sisteminin çalışacağı sunucu
		ChannelPrefix:     "destek-",
		OpenButtonID:      "ticket_open",
		CloseButtonID:     "ticket_close",
		MaxTicketsPerUser: 1,
	})

	// 2. Buton Tıklamalarını (Interaction) Yakalama
	session.AddHandler(func(s *micup.Session, i *micup.InteractionCreate) {
		if i.Interaction == nil {
			return
		}
		// Ticket butonu tıklandıysa otomatik yönet
		handled, err := ticketMgr.HandleInteraction(i.Interaction)
		if err != nil {
			log.Printf("Ticket interaction hatası: %v", err)
		}
		if handled {
			return
		}
	})

	// 3. !ticket-kur Komutu ile Kanala Butonlu Panel Gönderme
	session.AddHandler(func(s *micup.Session, m *micup.MessageCreate) {
		if m.Author == nil || m.Author.Bot {
			return
		}

		if m.Content == "!ticket-kur" {
			embed, row := ticketMgr.CreatePanelMessage(
				"🎫 Destek ve Yardım Talebi",
				"Sunucu yetkililerimizle özel görüşmek ve yardım almak için aşağıdaki butona tıklayın.\nSize özel gizli bir destek kanalı açılacaktır.",
				"Destek Talebi Aç",
			)

			_, err := s.ChannelMessageSendComplex(m.ChannelID, &micup.MessageSend{
				Embeds:     []*micup.Embed{embed},
				Components: []micup.Component{row},
			})
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Panel gönderilemedi: %v", err))
				return
			}
			_, _ = s.ChannelMessageSend(m.ChannelID, "✅ Destek paneli başarıyla kuruldu!")
		}
	})

	// 4. Web Yönetim Panelini (Dashboard) Başlatma
	dashboard, err := session.StartDashboard(&micup.DashboardOptions{
		Host:  "0.0.0.0",
		Port:  8080,
		Title: "Micup Bot Yönetim Paneli",
	})
	if err != nil {
		log.Printf("Web paneli başlatılamadı: %v", err)
	} else {
		log.Printf("🌐 Web Dashboard yayında: http://localhost:8080 (veya http://%s)", dashboard.Addr())
	}

	// 5. Gateway Bağlantısını Aç
	if err := session.Open(); err != nil {
		log.Fatalf("Bağlantı hatası: %v", err)
	}
	defer session.Close()

	log.Println("Ticket & Dashboard Botu çalışıyor. Çıkmak için CTRL-C...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	if dashboard != nil {
		_ = dashboard.Stop()
	}
}
