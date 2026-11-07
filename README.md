# micup-go

[![Go Reference](https://pkg.go.dev/badge/github.com/Qazcto7/micup-go.svg)](https://pkg.go.dev/github.com/Qazcto7/micup-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**micup-go**, [Micup.gg](https://micup.gg) (Fluxer motoru tabanlı sesli ve metin iletişim platformu) için Go diliyle geliştirilmiş, yasal ve kurallara %100 tam uyumlu bot kütüphanesidir (SDK).

Bu kütüphane, [morluto/rea](https://github.com/morluto/rea) (Reverse Engineer Anything) araç seti ve tersine mühendislik ilkeleri doğrultusunda Micup masaüstü (Electron/asar) ve web protokolü analiz edilerek hazırlanmıştır.

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
- 🎙️ **Ses Durumu ve Bağlantı Yönetimi (`voice.go`):**
  - Ses kanalına katılma (`ChannelVoiceJoin`) ve ayrılma (`ChannelVoiceLeave`)
  - `VOICE_STATE_UPDATE` ve `VOICE_SERVER_UPDATE` yönlendirmesi
  - `VoiceConnection` ile ses kimlik doğrulama belirteçlerinin (token, endpoint) takibi
- 🎨 **Akıcı Embed Oluşturucu (`embed.go`):**
  - Zincirleme (Fluent) API: `NewEmbedBuilder().SetTitle(...).SetColor(...).AddField(...).Build()`
- 🔘 **Etkileşimli UI Bileşenleri (`components.go`):**
  - Eylem Satırları (`ActionRow`) ve Düğmeler (`Button` - Primary, Secondary, Success, Danger, Link)
- 🛡️ **Moderasyon & Denetim Kayıtları (`moderation.go`):**
  - Toplu mesaj silme (`ChannelMessagesBulkDelete` - 100 mesaja kadar tek seferde)
  - Yazıyor bildirimi (`ChannelTyping`)
  - Sunucu yasakları (`GuildBans`), üye budama (`GuildPrune`) ve denetim kayıtları (`GuildAuditLogs`)
- 🧵 **İş Parçacıkları & Forum Kanalları (`thread.go`):**
  - Metin veya mesajdan iş parçacığı başlatma (`ChannelThreadStart`, `MessageThreadStart`)
  - Aktif iş parçacıklarını listeleme (`ChannelActiveThreads`, `GuildActiveThreads`)
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
go get github.com/Qazcto7/micup-go@v1.1.0
```

---

## 🏁 Kod Örnekleri

### 1. EmbedBuilder ve Butonlu Mesaj Örneği

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

	router := micup.NewCommandRouter("!")

	router.Register(&micup.Command{
		Name:        "panel",
		Description: "Etkileşimli butonlu kontrol panelini açar.",
		Handler: func(ctx *micup.CommandContext) error {
			// Akıcı EmbedBuilder ile kart hazırla
			embed := micup.NewEmbedBuilder().
				SetTitle("🎮 Topluluk Kontrol Paneli").
				SetDescription("Aşağıdaki butonları kullanarak işlem yapabilirsiniz.").
				SetColor(0x5865F2).
				AddField("Sunucu", "Micup Resmi Topluluğu", true).
				AddField("Durum", "Aktif 🟢", true).
				SetFooter("Micup Bot", "").
				Build()

			// Butonları oluştur
			btnVote := micup.NewSuccessButton("btn_vote", "Oy Ver ⭐")
			btnWebsite := micup.NewLinkButton("https://micup.gg", "Siteyi Aç 🌐")
			row := micup.NewActionRow(btnVote, btnWebsite)

			// Gönder
			_, err := ctx.Session.ChannelMessageSendComplex(ctx.Message.ChannelID, &micup.MessageSend{
				Embeds:     []*micup.Embed{embed},
				Components: []micup.Component{row},
			})
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

### 2. Ses Kanalına Katılma Örneği (`voice.go`)

```go
sess.AddHandler(func(s *micup.Session, m *micup.MessageCreate) {
    if m.Content == "!katil" {
        // Ses kanalına katıl (guildID, channelID, selfMute, selfDeaf)
        vc, err := s.ChannelVoiceJoin(context.Background(), m.GuildID, "kanal-id-123", false, false)
        if err != nil {
            log.Printf("Ses kanalına bağlanılamadı: %v", err)
            return
        }
        log.Printf("Ses oturumu başlatıldı: Guild %s, Channel %s", vc.GuildID, vc.ChannelID)
    }

    if m.Content == "!ayril" {
        _ = s.ChannelVoiceLeave(m.GuildID)
    }
})
```

---

### 3. Toplu Mesaj Temizleme (Moderasyon)

```go
router.Register(&micup.Command{
    Name: "sil",
    Description: "Kanaldaki son mesajları topluca siler.",
    Handler: func(ctx *micup.CommandContext) error {
        messages, err := ctx.Session.REST().ChannelMessages(context.Background(), ctx.Message.ChannelID, 50, "", "", "")
        if err != nil {
            return err
        }
        var ids []string
        for _, msg := range messages {
            ids = append(ids, msg.ID)
        }
        return ctx.Session.ChannelMessagesBulkDelete(ctx.Message.ChannelID, ids)
    },
})
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
| **OpCode 3** | `PRESENCE_UPDATE` | Durum ve aktivite güncelleme |
| **OpCode 4** | `VOICE_STATE_UPDATE` | Ses kanalına katılma ve ayrılma |
| **OpCode 6** | `RESUME` | Kesinti sonrası oturum kurtarma paketi |
| **OpCode 10** | `HELLO` | `heartbeat_interval` bilgisi |
| **Voice Engine V2** | `@fluxer/voice_engine_v2` | Ses sunucusu ve LiveKit oturum el sıkışması |

---

## 📄 Lisans

Bu proje [MIT Lisansı](LICENSE) kapsamında lisanslanmıştır.
Micup, O Group MMC'nin tescilli ticari markasıdır. Bu kütüphane topluluk odaklı, yasal ve açık kaynaklı bağımsız bir projedir.
