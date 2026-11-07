// Package micup provides an idiomatic, full-featured, and compliant Go SDK for building
// bots on the Micup.gg platform (based on the Fluxer engine).
//
// Features include:
//   - Automatic dynamic discovery via /.well-known/fluxer
//   - Resilient real-time WebSocket Gateway client with heartbeat, reconnect, and session resume
//   - Fully typed Gateway dispatch events and concurrency-safe handler registry
//   - Robust REST client with rate limiting and exponential backoff
//   - Channel and message management (text, embeds, reactions, pins, pagination)
//   - Server/Guild administration (roles, channels, members, moderation)
//   - Built-in extensible message Command Router
//   - MicupList (api.micuplist.me) ecosystem integration (server stats, vote checking, webhooks)
//   - 100% compliant with Micup developer terms and legal standards
package micup
