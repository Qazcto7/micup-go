# micup-go

[![Go Reference](https://pkg.go.dev/badge/github.com/Qazcto7/micup-go.svg)](https://pkg.go.dev/github.com/Qazcto7/micup-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/Qazcto7/micup-go/actions/workflows/ci.yml/badge.svg)](https://github.com/Qazcto7/micup-go/actions/workflows/ci.yml)

**micup-go**, [Micup.gg](https://micup.gg) (Fluxer motoru tabanlı sesli ve metin iletişim platformu) için Go diliyle geliştirilmiş, yasal ve kurallara %100 tam uyumlu, kurumsal seviyede bot SDK'sıdır.

Bu kütüphane, [morluto/rea](https://github.com/morluto/rea) (Reverse Engineer Anything) ve [Ghidra](https://ghidra-sre.org/) tersine mühendislik araçlarıyla Micup istemci çekirdeği (`app.asar`), ses modülleri (`linux-audio-capture`, `hardware-encoder`) ve protokol katmanı analiz edilerek hazırlanmıştır.

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

- 🎫 **Butonlu Destek & Ticket Sistemi (`ticket.go`):**
  - Butonlu etkileşim paneli (`CreatePanelMessage`).
  - Kullanıcıya ve yetkili rollerine özel izole destek kanalı açma (`PermissionOverwrite`).
  - Kapatma butonu, loglama kanalı ve otomatik kanal silme (`TicketManager`).
- 🌐 **Dahili Web Yönetim Paneli (Dashboard - `dashboard.go`):**
  - Tek satırla ayağa kalkan web arayüzü (`s.StartDashboard(&DashboardOptions{Port: 8080})`).
  - Gerçek zamanlı sunucu, üye, kanal istatistikleri ve uptime takibi.
  - Canlı Gateway olay akışı logları ve tarayıcıdan kanallara hızlı mesaj gönderme formu.
- 🧠 **Hafıza-içi State Önbelleği (`state.go`):**
  - Sunucular (`Guilds`), Kanallar (`Channels`), Üyeler (`Members`), Roller (`Roles`) ve mesaj geçmişi otomatik olarak RAM'de güncel tutulur.
  - Gereksiz REST API isteklerini önler; `s.State.Guild(id)`, `s.State.Channel(id)` ile mikrosaniyeler içinde erişim sağlar.
- 👑 **Bitwise İzin & Yetki Hesaplayıcı (`permissions.go`):**
  - 64-bit izin bitmask sabitleri (`PermissionAdministrator`, `PermissionManageGuild`, `PermissionSendMessages`, vb.).
  - `@everyone`, üye rolleri ve kanal bazlı `PermissionOverwrite` (Allow/Deny) izin ezmelerini çözen `ComputeMemberPermissions` motoru.
- 🎵 **FFmpeg Opus Ses Yayıncısı & Çalar (`audio_player.go` & `voice.go`):**
  - LiveKit WebRTC ses kanallarına bağlanma (`VoiceJoin`).
  - FFmpeg ile her türlü ses dosyasını veya canlı yayını anında 48kHz Stereo 20ms PCM paketlerine dönüştürüp akıtma.
  - Gerçek zamanlı ses düzeyi ayarlama (`SetVolume`), duraklatma (`Pause`), devam ettirme (`Resume`) ve durdurma (`Stop`).
- 📹 **Canlı Video & Ekran Yayını Motoru (`video_stream.go` & `voice.go`):**
  - Ses odalarında kamera (`camera`) veya ekran paylaşımı (`screen`) video kanalları açma (`PublishVideoTrack`).
  - FFmpeg ile Annex-B H.264 NAL birimlerini 30/60 FPS zamanlayıcıyla canlı yayınlama (`VideoStreamer`, `SendVideoFrame`).
- ⚡ **Donanım Video Kodlayıcı Tespiti (`encoder_probe.go`):**
  - NVIDIA NVENC, Intel/AMD VA-API ve FFmpeg donanım video kodlayıcılarını dinamik sorgulama (`ProbeHardwareEncoders`).
- 🎮 **Oyun Aktivitesi & Süreç Tanıma Protokolü (`game_activity.go`):**
  - Micup istemcisinin (`MainApp.js` & `preload`) kullandığı `game_id`, `manual_name` ve `started_at` standartlarına %100 uyumlu oyun durumu yayınlama (`s.UpdatePlayingGame`).
- 🛡️ **Otomatik Moderasyon Motoru (`automod.go`):**
  - Platform REST AutoMod kuralları yönetimi.
  - Gerçek zamanlı bot içi AutoMod middleware'i (`AutoModEngine`): Regex kalıpları, yasaklı kelimeler ve mention spam koruması.
- 📋 **Denetim Kayıtları API'si (`audit_log.go`):**
  - Sunucudaki ban, kick, rol, kanal silme/oluşturma gibi tüm yönetici eylemlerini listeleme ve filtreleme (`GuildAuditLogs`).
- 😀 **Özel Emoji & Mesaj Reaksiyonları (`emoji.go`):**
  - Base64 ile sunucuya özel emoji yükleme, düzenleme ve silme.
  - Mesajlara reaksiyon ekleme (`MessageReactionAdd`), silme ve reaksiyon veren kullanıcıları listeleme.
- 📅 **Planlanmış Sunucu Etkinlikleri (`scheduled_event.go`):**
  - Sesli, sahne ve harici etkinlik oluşturma (`GuildScheduledEventCreate`), güncelleme ve katılımcı takibi.
- 🌐 **Otomatik Dinamik Keşif (`discovery.go`):** Canlı Gateway ve REST API adreslerini `/.well-known/fluxer` üzerinden otomatik tespit eder.
- ⚡ **WebSocket Gateway v1 (`gateway.go`):**
  - Kalp Atışı (Heartbeat) & Gecikme (Jitter) yönetimi.
  - Otomatik Oturum Kurtarma (`RESUME`) ve Yeniden Bağlanma (`Exponential Backoff`).
  - Thread-safe (goroutine yarışsız) tip güvenli olay dinleme mimarisi.
- 🗂️ **Gateway Sharding (`sharding.go`):** Büyük ölçekli botlar için birden fazla shard'ı koordine eden `ShardingManager`.
- 📁 **Dosya & Medya Yükleme (`file.go`):** `multipart/form-data` ile resim, video ve belge gönderme (`ChannelMessageSendFile`, `ChannelMessageSendFiles`).
- ⚡ **Slash Commands & Interactions (`interaction.go`):**
  - Global ve sunucuya özel Application Command kaydı (`CreateGlobalApplicationCommand`).
  - `INTERACTION_CREATE` yakalama ve yanıt gönderme (`InteractionRespond`, `RespondText`).
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
- 🎨 **Akıcı Embed Oluşturucu (`embed.go`):**
  - Zincirleme (Fluent) API: `NewEmbed().SetTitle(...).SetColor(...).AddField(...).Build()`.
- 🧵 **İş Parçacıkları & Forum Kanalları (`thread.go`):**
  - Metin veya mesajdan iş parçacığı başlatma (`ChannelThreadStart`, `MessageThreadStart`).
- 🛠️ **Gelişmiş Komut Yönlendiricisi (`command.go`):**
  - Özelleştirilebilir komut öneki (`!`, `/`, `?` vb.) ve otomatik `!help` yardım menüsü.
- 🌟 **MicupList (api.micuplist.me) Entegrasyonu (`micuplist.go`):**
  - Bot sunucu sayısı istatistiğini otomatik gönderme ve canlı oy webhook dinleyicisi.

---

## 📦 Kurulum

```bash
go get github.com/Qazcto7/micup-go@v1.6.0
```

---

## 🏁 Kod Örnekleri

### 1. Canlı Ses Kanalında Müzik/Ses Oynatma (`AudioPlayer`)

```go
package main

import (
    "github.com/Qazcto7/micup-go"
)

func main() {
    client, _ := micup.New("TOKEN")

    client.AddHandler(func(s *micup.Session, m *micup.MessageCreate) {
        if m.Content == "!çal" {
            // Ses kanalına bağlan
            vc, _ := s.VoiceJoin(m.GuildID, "KANAL_ID", false, false)
            
            // FFmpeg ses oynatıcısı başlat
            player := vc.NewPlayer()
            player.SetVolume(0.8) // %80 ses
            _ = player.Play("muzik.mp3") // veya doğrudan URL
        }
    })

    _ = client.Open()
    defer client.Close()
    select {}
}
```

### 2. Otomatik Moderasyon & Filtreleme (`AutoModEngine`)

```go
automod := micup.NewAutoModEngine()

// Yasaklı kelime filtresi
automod.AddKeywordRule("reklam", []string{"discord.gg/", "t.me/"}, func(s *micup.Session, msg *micup.Message, reason string) {
    _ = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
    _, _ = s.ChannelMessageSend(msg.ChannelID, "⚠️ Reklam içerikleri engellendi!")
})

// Aşırı etiket (mass-mention) engeli
automod.AddMentionLimitRule("mass_mention", 5, func(s *micup.Session, msg *micup.Message, reason string) {
    _ = s.ChannelMessageDelete(msg.ChannelID, msg.ID)
    _, _ = s.ChannelMessageSend(msg.ChannelID, "🚨 Çok fazla etiketleme yapamazsınız!")
})

automod.Attach(session)
```

### 3. Etkileşimli Butonlar ve Component Router

```go
compRouter := micup.NewComponentRouter()
compRouter.Handle("btn_approve", func(s *micup.Session, i *micup.Interaction) {
    _ = s.RespondText(i, "✅ İşlem onaylandı!")
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
```

---

## 📄 Lisans

Bu proje [MIT Lisansı](LICENSE) kapsamında lisanslanmıştır.
Micup, O Group MMC'nin tescilli ticari markasıdır. Bu kütüphane topluluk odaklı, yasal ve açık kaynaklı bağımsız bir projedir.
