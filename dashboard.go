package micup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// DashboardOptions specifies configuration for the Web Dashboard server.
type DashboardOptions struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Title    string `json:"title"`
	AdminKey string `json:"admin_key,omitempty"`
}

// DashboardLog represents an event logged on the web dashboard.
type DashboardLog struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Message   string `json:"message"`
}

// Dashboard provides a browser-based management interface for the bot.
type Dashboard struct {
	session   *Session
	options   *DashboardOptions
	server    *http.Server
	startTime time.Time
	logs      []DashboardLog
	mu        sync.RWMutex
}

// NewDashboard initializes the web dashboard server.
func NewDashboard(s *Session, opts *DashboardOptions) *Dashboard {
	if opts == nil {
		opts = &DashboardOptions{}
	}
	if opts.Host == "" {
		opts.Host = "0.0.0.0"
	}
	if opts.Port <= 0 {
		opts.Port = 8080
	}
	if opts.Title == "" {
		opts.Title = "Micup Bot Dashboard"
	}

	d := &Dashboard{
		session:   s,
		options:   opts,
		startTime: time.Now(),
		logs:      make([]DashboardLog, 0),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", d.handleIndex)
	mux.HandleFunc("/api/stats", d.handleStats)
	mux.HandleFunc("/api/guilds", d.handleGuilds)
	mux.HandleFunc("/api/send", d.handleSendMessage)
	mux.HandleFunc("/api/logs", d.handleLogs)

	d.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", opts.Host, opts.Port),
		Handler: mux,
	}

	// Auto-hook into session event dispatch to record logs
	s.AddHandler(func(s *Session, e *Ready) {
		d.AddLog("READY", fmt.Sprintf("Bot oturumu hazırlandı: %s", s.State.User.Username))
	})
	s.AddHandler(func(s *Session, e *GuildCreate) {
		d.AddLog("GUILD_CREATE", fmt.Sprintf("Sunucu yüklendi: %s (%s)", e.Name, e.ID))
	})
	s.AddHandler(func(s *Session, e *MessageCreate) {
		if e.Author != nil {
			d.AddLog("MESSAGE_CREATE", fmt.Sprintf("<#%s> - %s: %s", e.ChannelID, e.Author.Username, e.Content))
		}
	})

	return d
}

// AddLog pushes an entry into the dashboard log buffer.
func (d *Dashboard) AddLog(logType, message string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	entry := DashboardLog{
		Timestamp: time.Now().Format("15:04:05"),
		Type:      logType,
		Message:   message,
	}

	d.logs = append(d.logs, entry)
	if len(d.logs) > 200 {
		d.logs = d.logs[1:]
	}
}

// Start launches the HTTP server in a separate background goroutine.
func (d *Dashboard) Start() error {
	d.AddLog("SYSTEM", fmt.Sprintf("Web Dashboard başlatıldı: http://%s:%d", d.options.Host, d.options.Port))
	go func() {
		_ = d.server.ListenAndServe()
	}()
	return nil
}

// Stop gracefully shuts down the HTTP dashboard server.
func (d *Dashboard) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return d.server.Shutdown(ctx)
}

// Addr returns the listening network address.
func (d *Dashboard) Addr() string {
	return d.server.Addr
}

func (d *Dashboard) handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	uptime := time.Since(d.startTime)
	guilds := d.session.State.GuildsList()
	totalMembers := 0
	totalChannels := 0

	for _, g := range guilds {
		totalMembers += g.MemberCount
		totalChannels += len(g.Channels)
	}

	username := "Micup Bot"
	userID := "0"
	if d.session.State.User != nil {
		username = d.session.State.User.Username
		userID = d.session.State.User.ID
	}

	stats := map[string]interface{}{
		"bot": map[string]string{
			"id":       userID,
			"username": username,
		},
		"uptime_seconds":   int(uptime.Seconds()),
		"uptime_formatted": formatDuration(uptime),
		"guild_count":      len(guilds),
		"channel_count":    totalChannels,
		"member_count":     totalMembers,
		"title":            d.options.Title,
	}

	_ = json.NewEncoder(w).Encode(stats)
}

func (d *Dashboard) handleGuilds(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	guilds := d.session.State.GuildsList()
	_ = json.NewEncoder(w).Encode(guilds)
}

func (d *Dashboard) handleLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	d.mu.RLock()
	defer d.mu.RUnlock()
	_ = json.NewEncoder(w).Encode(d.logs)
}

func (d *Dashboard) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ChannelID string `json:"channel_id"`
		Content   string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ChannelID == "" || req.Content == "" {
		http.Error(w, "Geçersiz istek parametreleri", http.StatusBadRequest)
		return
	}

	_, err := d.session.ChannelMessageSend(req.ChannelID, req.Content)
	if err != nil {
		http.Error(w, fmt.Sprintf("Mesaj gönderilemedi: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02ds %02dd %02dsn", h, m, s)
}

// Embedded Single Page Dark-Themed Web UI
func (d *Dashboard) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashboardHTML))
}

// StartDashboard is a convenience method on Session to launch the Web Dashboard.
func (s *Session) StartDashboard(opts *DashboardOptions) (*Dashboard, error) {
	d := NewDashboard(s, opts)
	if err := d.Start(); err != nil {
		return nil, err
	}
	return d, nil
}

const dashboardHTML = `<!DOCTYPE html>
<html lang="tr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Micup Bot Dashboard</title>
    <style>
        :root {
            --bg-dark: #0f172a;
            --bg-card: #1e293b;
            --bg-hover: #334155;
            --text-main: #f8fafc;
            --text-muted: #94a3b8;
            --accent: #3b82f6;
            --accent-green: #10b981;
            --border: #334155;
        }
        * { box-sizing: border-box; margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; }
        body { background: var(--bg-dark); color: var(--text-main); padding: 24px; }
        .container { max-width: 1200px; margin: 0 auto; }
        header { display: flex; justify-content: space-between; align-items: center; padding-bottom: 24px; border-bottom: 1px solid var(--border); margin-bottom: 24px; }
        .bot-badge { display: flex; align-items: center; gap: 12px; }
        .avatar { width: 48px; height: 48px; border-radius: 50%; background: var(--accent); display: flex; align-items: center; justify-content: center; font-weight: bold; font-size: 20px; }
        .status-dot { display: inline-block; width: 10px; height: 10px; background: var(--accent-green); border-radius: 50%; margin-right: 6px; }
        .stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(240px, 1fr)); gap: 16px; margin-bottom: 24px; }
        .stat-card { background: var(--bg-card); border: 1px solid var(--border); border-radius: 12px; padding: 20px; }
        .stat-card h3 { font-size: 14px; color: var(--text-muted); text-transform: uppercase; margin-bottom: 8px; }
        .stat-card .value { font-size: 28px; font-weight: 700; color: #fff; }
        .section-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 24px; }
        @media (max-width: 800px) { .section-grid { grid-template-columns: 1fr; } }
        .card { background: var(--bg-card); border: 1px solid var(--border); border-radius: 12px; padding: 20px; }
        .card h2 { font-size: 18px; margin-bottom: 16px; border-bottom: 1px solid var(--border); padding-bottom: 8px; display: flex; justify-content: space-between; }
        .log-list { height: 260px; overflow-y: auto; font-family: monospace; font-size: 13px; background: #0b1120; padding: 12px; border-radius: 8px; }
        .log-item { margin-bottom: 6px; line-height: 1.4; }
        .log-time { color: var(--text-muted); }
        .log-type { color: var(--accent); font-weight: bold; }
        .form-group { margin-bottom: 14px; }
        label { display: block; font-size: 13px; color: var(--text-muted); margin-bottom: 6px; }
        input, select, textarea { width: 100%; background: #0b1120; border: 1px solid var(--border); border-radius: 6px; padding: 10px; color: #fff; font-size: 14px; }
        button { background: var(--accent); border: none; border-radius: 6px; padding: 10px 18px; color: #fff; font-weight: 600; cursor: pointer; transition: 0.2s; }
        button:hover { background: #2563eb; }
        .guild-item { display: flex; justify-content: space-between; padding: 10px 0; border-bottom: 1px solid rgba(255,255,255,0.05); }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <div class="bot-badge">
                <div class="avatar">🤖</div>
                <div>
                    <h1 id="bot-name">Micup Bot</h1>
                    <p style="color: var(--text-muted); font-size: 14px;"><span class="status-dot"></span>Çevrimiçi • <span id="uptime">0s</span></p>
                </div>
            </div>
            <button onclick="fetchStats()">🔄 Yenile</button>
        </header>

        <div class="stats-grid">
            <div class="stat-card">
                <h3>Sunucu Sayısı</h3>
                <div class="value" id="stat-guilds">0</div>
            </div>
            <div class="stat-card">
                <h3>Kanal Sayısı</h3>
                <div class="value" id="stat-channels">0</div>
            </div>
            <div class="stat-card">
                <h3>Kullanıcı Sayısı</h3>
                <div class="value" id="stat-members">0</div>
            </div>
            <div class="stat-card">
                <h3>Uptime</h3>
                <div class="value" id="stat-uptime">-</div>
            </div>
        </div>

        <div class="section-grid">
            <div class="card">
                <h2>Sunucular</h2>
                <div id="guild-list" style="max-height: 280px; overflow-y: auto;">
                    <p style="color: var(--text-muted);">Sunucular yükleniyor...</p>
                </div>
            </div>

            <div class="card">
                <h2>Hızlı Mesaj Gönder</h2>
                <div class="form-group">
                    <label>Kanal ID</label>
                    <input type="text" id="send-channel" placeholder="Örn: 1568205528151625740">
                </div>
                <div class="form-group">
                    <label>Mesaj İçeriği</label>
                    <textarea id="send-content" rows="3" placeholder="Gönderilecek duyuru veya mesaj..."></textarea>
                </div>
                <button onclick="sendMessage()">Mesajı Gönder 🚀</button>
                <span id="send-status" style="margin-left: 10px; font-size: 13px;"></span>
            </div>
        </div>

        <div class="card" style="margin-top: 24px;">
            <h2>Canlı Olay Logları (Gateway Events)</h2>
            <div class="log-list" id="logs-container">
                <p style="color: var(--text-muted);">Loglar dinleniyor...</p>
            </div>
        </div>
    </div>

    <script>
        async function fetchStats() {
            try {
                const res = await fetch('/api/stats');
                const data = await res.json();
                document.getElementById('bot-name').innerText = data.bot.username;
                document.getElementById('stat-guilds').innerText = data.guild_count;
                document.getElementById('stat-channels').innerText = data.channel_count;
                document.getElementById('stat-members').innerText = data.member_count;
                document.getElementById('stat-uptime').innerText = data.uptime_formatted;
                document.getElementById('uptime').innerText = data.uptime_formatted;
            } catch (e) {
                console.error(e);
            }
        }

        async function fetchGuilds() {
            try {
                const res = await fetch('/api/guilds');
                const guilds = await res.json();
                const container = document.getElementById('guild-list');
                if (!guilds || guilds.length === 0) {
                    container.innerHTML = '<p style="color: var(--text-muted);">Henüz sunucu bulunmuyor.</p>';
                    return;
                }
                container.innerHTML = guilds.map(function(g) {
                    return '<div class="guild-item">' +
                        '<div>' +
                            '<strong>' + g.name + '</strong>' +
                            '<div style="font-size: 12px; color: var(--text-muted);">' + g.id + '</div>' +
                        '</div>' +
                        '<div style="text-align: right; color: var(--accent); font-size: 14px;">' +
                            (g.member_count || 0) + ' üye' +
                        '</div>' +
                    '</div>';
                }).join('');
            } catch (e) {
                console.error(e);
            }
        }

        async function fetchLogs() {
            try {
                const res = await fetch('/api/logs');
                const logs = await res.json();
                const container = document.getElementById('logs-container');
                if (logs && logs.length > 0) {
                    container.innerHTML = logs.map(function(l) {
                        return '<div class="log-item">' +
                            '<span class="log-time">[' + l.timestamp + ']</span> ' +
                            '<span class="log-type">[' + l.type + ']</span> ' +
                            '<span>' + l.message + '</span>' +
                        '</div>';
                    }).join('');
                    container.scrollTop = container.scrollHeight;
                }
            } catch (e) {}
        }

        async function sendMessage() {
            const ch = document.getElementById('send-channel').value.trim();
            const txt = document.getElementById('send-content').value.trim();
            const status = document.getElementById('send-status');

            if (!ch || !txt) {
                status.style.color = '#ef4444';
                status.innerText = 'Lütfen kanal ID ve mesaj girin!';
                return;
            }

            try {
                const res = await fetch('/api/send', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ channel_id: ch, content: txt })
                });
                if (res.ok) {
                    status.style.color = '#10b981';
                    status.innerText = '✅ Gönderildi!';
                    document.getElementById('send-content').value = '';
                    setTimeout(() => status.innerText = '', 3000);
                } else {
                    const err = await res.text();
                    status.style.color = '#ef4444';
                    status.innerText = 'Hata: ' + err;
                }
            } catch (e) {
                status.style.color = '#ef4444';
                status.innerText = 'İstek başarısız';
            }
        }

        fetchStats();
        fetchGuilds();
        fetchLogs();
        setInterval(fetchStats, 5000);
        setInterval(fetchLogs, 3000);
    </script>
</body>
</html>
`
