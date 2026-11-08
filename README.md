# micup-go

[![Go Reference](https://pkg.go.dev/badge/github.com/Qazcto7/micup-go.svg)](https://pkg.go.dev/github.com/Qazcto7/micup-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/Qazcto7/micup-go/actions/workflows/ci.yml/badge.svg)](https://github.com/Qazcto7/micup-go/actions/workflows/ci.yml)

**micup-go**, [Micup.gg](https://micup.gg) (Fluxer motoru tabanlı sesli ve metin iletişim platformu) için Go diliyle geliştirilmiş, yasal ve kurallara %100 tam uyumlu, kurumsal seviyede bot SDK'sıdır.

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

## 🚀 Kapsamlı Özellikler

- 🌐 **Otomatik Dinamik Keşif (`discovery.go`):** Canlı Gateway ve REST API adreslerini `/.well-known/fluxer` üzerinden otomatik tespit eder.
- ⚡ **WebSocket Gateway v1 (`gateway.go`):**
  - Kalp Atışı (Heartbeat) & Gecikme (Jitter) yönetimi
  - Otomatik Oturum Kurtarma (`RESUME`) ve Yeniden Bağlanma (`Exponential Backoff`)
  - Tüm Gateway OpCode'ları (`DISPATCH`, `HEARTBEAT`, `IDENTIFY`, `RESUME`, `RECONNECT`, `INVALID_SESSION`, `HELLO`, vb.)
  - Tip güvenli olay (event) dinleme mimarisi
- 🗂️ **Gateway Sharding (`sharding.go`):** Büyük ölçekli botlar için birden fazla shard'ı koordine eden `ShardingManager`.
- 📁 **Dosya & Medya Yükleme (`file.go`):** `multipart/form-data` ile resim, video ve belge gönderme (`ChannelMessageSendFile`, `ChannelMessageSendFiles`).
- ⚡ **Slash Commands & Interactions (`interaction.go`):**
  - Global ve sunucuya özel Application Command kaydı (`CreateGlobalApplicationCommand`)
  - `INTERACTION_CREATE` yakalama ve yanıt gönderme (`InteractionRespond`, `RespondText`)
- 🎙️ **Ses Durumu ve Canlı Yayın (`voice.go` & `voice_stream.go`):**
  - Ses kanalına katılma (`ChannelVoiceJoin`) ve ayrılma (`ChannelVoiceLeave`)
  - Canlı Opus ses akışı (`AudioStreamer`, `SendOpusFrame`, `StreamOpusReader`)
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
- 🛠️ **Gelişmiş Komut Yönlendiricisi (`command.go`):**
  - Özelleştirilebilir komut öneki (`!`, `/`, `?` vb.)
  - Otomatik `!help` yardım menüsü
  - Kolaylaştırılmış `ctx.Reply()` ve `ctx.ReplyEmbed()` metodları
- 🌟 **MicupList (api.micuplist.me) Entegrasyonu (`micuplist.go`):**
  - Bot sunucu sayısı istatistiğini otomatik gönderme
  - Kullanıcı oy (vote) kontrolü (`CheckUserVote`)
  - Canlı oy webhook dinleyicisi (`NewVoteWebhookHandler`)

---

## 📦 Kurulum

```bash
go get github.com/Qazcto7/micup-go@v1.2.0
```

---

## 🏁 Kod Örnekleri

### 1. Slash Command (Application Command) ve Yanıt

```go
sess.AddHandler(func(s *micup.Session, i *micup.InteractionCreate) {
    if i.Type == micup.InteractionApplicationCommand {
        if i.Data.Name == "ping" {
            _ = s.RespondText(i.Interaction, "Pong! 🏓 Slash komutu çalışıyor!")
        }
    }
})
```

### 2. Dosya Gönderme (`file.go`)

```go
file, err := os.Open("screenshot.png")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

_, err = sess.ChannelMessageSendFile(channelID, "screenshot.png", file, "İşte ekran görüntüsü!")
```

### 3. Çoklu Shard Yönetimi (`sharding.go`)

```go
// 4 Shard ile dağıtık bot başlatma
mgr, err := micup.NewShardingManager(os.Getenv("MICUP_BOT_TOKEN"), 4)
if err != nil {
    log.Fatal(err)
}

mgr.AddHandler(func(s *micup.Session, m *micup.MessageCreate) {
    if m.Content == "!shard" {
        _, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Shard: %d / %d", s.Options().ShardID, s.Options().ShardCount))
    }
})

_ = mgr.Open(context.Background())
```

---

## 📄 Lisans

Bu proje [MIT Lisansı](LICENSE) kapsamında lisanslanmıştır.
Micup, O Group MMC'nin tescilli ticari markasıdır. Bu kütüphane topluluk odaklı, yasal ve açık kaynaklı bağımsız bir projedir.
