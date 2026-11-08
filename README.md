# micup-go

[![Go Reference](https://pkg.go.dev/badge/github.com/Qazcto7/micup-go.svg)](https://pkg.go.dev/github.com/Qazcto7/micup-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/Qazcto7/micup-go/actions/workflows/ci.yml/badge.svg)](https://github.com/Qazcto7/micup-go/actions/workflows/ci.yml)

**micup-go**, [Micup.gg](https://micup.gg) (Fluxer motoru tabanlı sesli ve metin iletişim platformu) için Go diliyle geliştirilmiş, yasal ve kurallara %100 tam uyumlu, kurumsal seviyede bot SDK'sıdır.

Bu kütüphane, [morluto/rea](https://github.com/morluto/rea) (Reverse Engineer Anything) araç seti ve tersine mühendislik ilkeleri doğrultusunda Micup istemci çekirdeği (`app.asar`) analiz edilerek hazırlanmıştır.

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

## 🚀 Kapsamlı Özellikler

- 🧠 **Hafıza-içi State Önbelleği (`state.go`):**
  - Sunucular (`Guilds`), Kanallar (`Channels`), Üyeler (`Members`), Roller (`Roles`) ve mesaj geçmişi otomatik olarak RAM'de güncel tutulur.
  - Gereksiz REST API isteklerini önler; `s.State.Guild(id)`, `s.State.Channel(id)` ile mikrosaniyeler içinde erişim sağlar.
- 🌐 **Otomatik Dinamik Keşif (`discovery.go`):** Canlı Gateway ve REST API adreslerini `/.well-known/fluxer` üzerinden otomatik tespit eder.
- ⚡ **WebSocket Gateway v1 (`gateway.go`):**
  - Kalp Atışı (Heartbeat) & Gecikme (Jitter) yönetimi
  - Otomatik Oturum Kurtarma (`RESUME`) ve Yeniden Bağlanma (`Exponential Backoff`)
  - Tüm Gateway OpCode'ları (`DISPATCH`, `HEARTBEAT`, `IDENTIFY`, `RESUME`, `RECONNECT`, `INVALID_SESSION`, `HELLO`, vb.)
  - Thread-safe (goroutine yarışsız) tip güvenli olay dinleme mimarisi.
- 🗂️ **Gateway Sharding (`sharding.go`):** Büyük ölçekli botlar için birden fazla shard'ı koordine eden `ShardingManager`.
- 📁 **Dosya & Medya Yükleme (`file.go`):** `multipart/form-data` ile resim, video ve belge gönderme (`ChannelMessageSendFile`, `ChannelMessageSendFiles`).
- ⚡ **Slash Commands & Interactions (`interaction.go`):**
  - Global ve sunucuya özel Application Command kaydı (`CreateGlobalApplicationCommand`)
  - `INTERACTION_CREATE` yakalama ve yanıt gönderme (`InteractionRespond`, `RespondText`)
- 🔘 **Etkileşimli UI Bileşenleri & Yönlendirici (`components.go` & `collector.go`):**
  - `ComponentRouter` ile `custom_id` bazlı buton/menü tıklama yönlendirmesi.
  - Eylem Satırları (`ActionRow`) ve Düğmeler (`Button` - Primary, Secondary, Success, Danger, Link).
  - Sayfalandırılmış Embed Gezgini (`EmbedPaginator`).
- 🎮 **Zengin Varlık ve Durum Yönetimi (`presence.go`):**
  - `s.UpdateGameStatus()`, `s.UpdateCustomStatus()`, `s.UpdateListeningStatus()`, `s.UpdateStreamingStatus()`.
- 🔗 **Webhooks API (`webhook.go`):**
  - Webhook oluşturma, düzenleme, silme ve mesaj gönderme (`WebhookExecute`).
- 📖 **Otomatik Sayfalayıcı & İteratörler (`paginator.go`):**
  - Sayfalar dolusu mesajı veya üyeyi otomatik sayfalandırarak çeken `MessageIterator` ve `MemberIterator`.
- 🎙️ **Ses Durumu ve Canlı Yayın (`voice.go` & `voice_stream.go`):**
  - Ses kanalına katılma (`ChannelVoiceJoin`) ve ayrılma (`ChannelVoiceLeave`)
  - Canlı Opus ses akışı (`AudioStreamer`, `SendOpusFrame`, `StreamOpusReader`)
- 🎨 **Akıcı Embed Oluşturucu (`embed.go`):**
  - Zincirleme (Fluent) API: `NewEmbed().SetTitle(...).SetColor(...).AddField(...).Build()`
- 🛡️ **Moderasyon & Denetim Kayıtları (`moderation.go`):**
  - Toplu mesaj silme (`ChannelMessagesBulkDelete` - 100 mesaja kadar tek seferde)
  - Yazıyor bildirimi (`ChannelTyping`)
  - Sunucu yasakları (`GuildBans`), üye budama (`GuildPrune`) ve denetim kayıtları (`GuildAuditLogs`)
- 🧵 **İş Parçacıkları & Forum Kanalları (`thread.go`):**
  - Metin veya mesajdan iş parçacığı başlatma (`ChannelThreadStart`, `MessageThreadStart`)
  - Aktif iş parçacıklarını listeleme (`ChannelActiveThreads`, `GuildActiveThreads`)
- 🛠️ **Gelişmiş Komut Yönlendiricisi (`command.go`):**
  - Özelleştirilebilir komut öneki (`!`, `/`, `?` vb.)
  - Otomatik `!help` yardım menüsü
- 🌟 **MicupList (api.micuplist.me) Entegrasyonu (`micuplist.go`):**
  - Bot sunucu sayısı istatistiğini otomatik gönderme
  - Kullanıcı oy (vote) kontrolü (`CheckUserVote`)
  - Canlı oy webhook dinleyicisi (`NewVoteWebhookHandler`)

---

## 📦 Kurulum

```bash
go get github.com/Qazcto7/micup-go@v1.3.0
```

---

## 🏁 Kod Örnekleri

### 1. Etkileşimli Butonlar ve Component Router

```go
package main

import (
    "github.com/Qazcto7/micup-go"
)

func main() {
    client, _ := micup.New("TOKEN")

    compRouter := micup.NewComponentRouter()
    compRouter.Handle("btn_approve", func(s *micup.Session, i *micup.Interaction) {
        _ = s.RespondText(i, "✅ Onaylandı!")
    })
    client.AddHandler(compRouter.Listener())

    client.AddHandler(func(s *micup.Session, m *micup.MessageCreate) {
        if m.Content == "!onay" {
            btn := micup.NewSuccessButton("btn_approve", "Onayla").WithEmoji("✅")
            _, _ = s.ChannelMessageSendComplex(m.ChannelID, &micup.MessageSend{
                Content:    "İşlemi onaylıyor musunuz?",
                Components: []micup.Component{micup.NewActionRow(btn)},
            })
        }
    })

    _ = client.Open()
    defer client.Close()
    select {}
}
```

### 2. State Önbelleğinden Hızlı Veri Okuma

```go
client.AddHandler(func(s *micup.Session, m *micup.MessageCreate) {
    if m.Content == "!server" {
        // Ağ isteği atmadan RAM'den anında çeker
        guild, err := s.State.Guild(m.GuildID)
        if err == nil {
            s.ChannelMessageSend(m.ChannelID, "Sunucu Adı: " + guild.Name)
        }
    }
})
```

### 3. Zengin Durum & Aktivite Güncelleme

```go
client.AddHandler(func(s *micup.Session, r *micup.Ready) {
    // Çalıyor, İzliyor, Oynuyor veya Özel Durum belirleyin
    _ = s.UpdateGameStatus(micup.StatusOnline, "Micup.gg | !yardim")
})
```

---

## 📄 Lisans

Bu proje [MIT Lisansı](LICENSE) kapsamında lisanslanmıştır.
Micup, O Group MMC'nin tescilli ticari markasıdır. Bu kütüphane topluluk odaklı, yasal ve açık kaynaklı bağımsız bir projedir.
