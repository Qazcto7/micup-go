# micup-go

[![Go Reference](https://pkg.go.dev/badge/github.com/Qazcto7/micup-go.svg)](https://pkg.go.dev/github.com/Qazcto7/micup-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**micup-go**, [Micup.gg](https://micup.gg) (Fluxer motoru tabanlı sesli ve metin iletişim platformu) için Go diliyle geliştirilmiş, yasal ve kurallara %100 tam uyumlu bot kütüphanesidir (SDK).

Bu kütüphane, [morluto/rea](https://github.com/morluto/rea) (Reverse Engineer Anything) yaklaşımı ve tersine mühendislik ilkeleri doğrultusunda Micup istemci ve ağ protokolü analiz edilerek hazırlanmıştır.

---

## ⚖️ Yasal ve Güvenli Kullanım İlkeleri

`micup-go`, Micup platformunun resmi geliştirici ve kullanım şartlarına ([Micup Kullanım Şartları](https://micup.gg/yasal/kullanim-sartlari.html)) saygı gösterecek şekilde tasarlanmıştır:
- **Resmi Bot Kimlik Doğrulaması:** Yetkilendirme için standart `Authorization: Bot <token>` (veya Bearer) başlığı kullanılır.
- **Şeffaf User-Agent:** Ağ trafiğinde botunuzun adı, sürümü ve kaynak linki şeffaf olarak belirtilir.
- **Dinamik Keşif (Dynamic Discovery):** URL'ler sabit kodlanmaz; doğrudan platformun `/.well-known/fluxer` keşif manifestosu üzerinden canlı servis uçları çözümlenir.
- **Hız Sınırı (Rate Limiting) Uyumu:** `429 Too Many Requests` ve `X-RateLimit-*` başlıklarına tam uyumlu olup, sunucuya gereksiz yük bindirmemek için exponential backoff (üstel gecikme) uygular.
- **Akıllı Gateway Bağlantısı:** Gateway v1 protokolü, jitter'lı heartbeat döngüsü ve kopmalarda oturum kurtarma (`RESUME` - OpCode 6) desteğiyle sunucu tarafında gereksiz el sıkışmalarını önler.
- **Spam ve Zararlı Faaliyet Yasağı:** Self-bot veya kullanıcı hesabı taklidi yapmaz, yalnızca yetkilendirilmiş bot token'ları ile çalışır.

---

## 🚀 Özellikler

- 🌐 **Otomatik Dinamik Keşif:** Canlı Gateway ve REST API adreslerini `/.well-known/fluxer` üzerinden otomatik tespit eder.
- ⚡ **WebSocket Gateway v1:**
  - Kalp Atışı (Heartbeat) & Gecikme (Jitter) yönetimi
  - Otomatik Oturum Kurtarma (`RESUME`) ve Yeniden Bağlanma (`Exponential Backoff`)
  - Tüm Gateway OpCode'ları (`DISPATCH`, `HEARTBEAT`, `IDENTIFY`, `RESUME`, `RECONNECT`, `INVALID_SESSION`, `HELLO`, vb.)
  - Tip güvenli olay (event) dinleme mimarisi
- 💬 **Kapsamlı REST API Desteği:**
  - **Kullanıcılar:** `@me` profil sorgulama, kullanıcı bilgileri alma, avatar ve kullanıcı adı güncelleme.
  - **Sunucular (Guilds):** Sunucu detayları, kanallar, üyeler, roller, rol atama/çıkarma, ban ve kick işlemleri.
  - **Kanallar & Mesajlar:** Mesaj gönderme, yanıtlama, düzenleme, silme, sabitleme (pin), zengin gömülü kartlar (Embeds) ve emojili tepkiler (Reactions).
- 🛠️ **Gelişmiş Komut Yönlendiricisi (Command Router):**
  - Özelleştirilebilir komut öneki (`!`, `/`, `?` vb.)
  - Otomatik `!help` yardım menüsü
  - Kolaylaştırılmış `ctx.Reply()` ve `ctx.ReplyEmbed()` metodları
- 🌟 **MicupList (api.micuplist.me) Entegrasyonu:**
  - Bot sunucu sayısı istatistiğini otomatik gönderme
  - Kullanıcı oy (vote) kontrolü (`CheckUserVote`)
  - Canlı oy webhook dinleyicisi (`NewVoteWebhookHandler`)

---

## 📦 Kurulum

```bash
go get github.com/Qazcto7/micup-go
```

---

## 🏁 Hızlı Başlangıç

### 1. Basit Ping-Pong Botu

```go
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
	// Bot token'ınızı girin
	token := os.Getenv("MICUP_BOT_TOKEN")

	sess, err := micup.New(token)
	if err != nil {
		log.Fatal(err)
	}

	// Bot hazır olduğunda tetiklenir
	sess.AddHandler(func(s *micup.Session, r *micup.Ready) {
		log.Printf("Giriş yapıldı: %s#%s", r.User.Username, r.User.Discriminator)
		_ = s.UpdateStatus("online", &micup.GatewayActivity{
			Name: "!ping dinliyor",
			Type: 0,
		})
	})

	// Gelen mesajları dinle
	sess.AddHandler(func(s *micup.Session, m *micup.MessageCreate) {
		if m.Author.Bot {
			return // Kendi mesajlarını ve diğer botları yok say
		}

		if m.Content == "!ping" {
			_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Pong! 🏓 Merhaba %s!", m.Author.Username))
		}
	})

	// Gateway bağlantısını aç
	if err := sess.Open(); err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	log.Println("Bot çalışıyor. Çıkmak için CTRL+C tuşlayın.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc
}
```

---

### 2. Komut Çerçevesi (Command Router) ve Embed Kullanımı

```go
package main

import (
	"log"
	"os"
	"github.com/Qazcto7/micup-go"
)

func main() {
	sess, err := micup.New(os.Getenv("MICUP_BOT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	// '!' önekiyle komut yöneticisi oluştur
	router := micup.NewCommandRouter("!")

	// !bilgi komutu
	router.Register(&micup.Command{
		Name:        "bilgi",
		Description: "Bot ve sunucu hakkında bilgi verir.",
		Handler: func(ctx *micup.CommandContext) error {
			embed := &micup.Embed{
				Title:       "Micup Bot Bilgisi",
				Description: "Bu bot micup-go kütüphanesiyle çalışıyor.",
				Color:       0x00D1B2,
				Fields: []*micup.EmbedField{
					{Name: "Kullanıcı", Value: ctx.Message.Author.Username, Inline: true},
					{Name: "Kanal ID", Value: ctx.Message.ChannelID, Inline: true},
				},
				Footer: &micup.EmbedFooter{Text: "Micup Topluluğu"},
			}
			_, err := ctx.ReplyEmbed(embed)
			return err
		},
	})

	router.Attach(sess)

	if err := sess.Open(); err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	select {}
}
```

---

### 3. MicupList (api.micuplist.me) Oy & İstatistik Entegrasyonu

```go
package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/Qazcto7/micup-go"
)

func main() {
	sess, _ := micup.New(os.Getenv("MICUP_BOT_TOKEN"))
	listClient := micup.NewMicupListClient(os.Getenv("MICUPLIST_API_TOKEN"))

	// Bot açılınca sunucu sayısını MicupList'e bildir
	sess.AddHandler(func(s *micup.Session, r *micup.Ready) {
		_ = listClient.PostStats(context.Background(), r.User.ID, len(r.Guilds))
	})

	// Oy veren kullanıcıları karşılayan webhook
	http.HandleFunc("/webhook/vote", micup.NewVoteWebhookHandler(os.Getenv("WEBHOOK_SECRET"), func(p *micup.VoteWebhookPayload) {
		log.Printf("Kullanıcı %s bota oy verdi!", p.UserID)
	}))
	go http.ListenAndServe(":8080", nil)

	_ = sess.Open()
	defer sess.Close()
	select {}
}
```

---

## 🔍 Tersine Mühendislik (REA) Bulguları Özeti

| Katman | Uç Nokta / Protokol | Açıklama |
|---|---|---|
| **Discovery** | `GET /.well-known/fluxer` | Dinamik API ve gateway uç noktalarını döndüren JSON |
| **REST API** | `https://micup.gg/api` | Kaynak tabanlı REST arayüzü (`/users`, `/guilds`, `/channels`, `/messages`) |
| **Gateway** | `wss://micup.gg/gateway?v=1&encoding=json` | Gerçek zamanlı WebSocket protokolü |
| **OpCode 0** | `DISPATCH` | READY, MESSAGE_CREATE, GUILD_CREATE vb. olaylar |
| **OpCode 1 / 11** | `HEARTBEAT` / `HEARTBEAT_ACK` | Çift yönlü canlı tutma sinyali |
| **OpCode 2** | `IDENTIFY` | Oturum açma paketi (`token`, `properties`, `presence`) |
| **OpCode 6** | `RESUME` | Kesinti sonrası oturum kurtarma paketi |
| **OpCode 10** | `HELLO` | `heartbeat_interval` bilgisi |

---

## 📄 Lisans

Bu proje [MIT Lisansı](LICENSE) kapsamında lisanslanmıştır.
Micup, O Group MMC'nin tescilli ticari markasıdır. Bu kütüphane topluluk odaklı, yasal ve açık kaynaklı bağımsız bir projedir.
