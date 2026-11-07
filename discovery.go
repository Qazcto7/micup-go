package micup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// WellKnownURL is the official discovery endpoint path for Fluxer / Micup instances.
const (
	DefaultBaseURL     = "https://micup.gg"
	DefaultWellKnown   = "/.well-known/fluxer"
	DefaultUserAgent   = "MicupBot (https://github.com/Qazcto7/micup-go, v1.0.0)"
)

// DiscoveryManifest represents the metadata returned by `/.well-known/fluxer`.
type DiscoveryManifest struct {
	APICodeVersion int                `json:"api_code_version"`
	Endpoints      DiscoveryEndpoints `json:"endpoints"`
	Captcha        DiscoveryCaptcha   `json:"captcha"`
	Features       DiscoveryFeatures  `json:"features"`
	Push           DiscoveryPush      `json:"push"`
	Legal          DiscoveryLegal     `json:"legal"`
	Limits         DiscoveryLimits    `json:"limits"`
}

// DiscoveryEndpoints contains all resolved service URLs for the platform.
type DiscoveryEndpoints struct {
	API          string `json:"api"`
	APIClient    string `json:"api_client"`
	APIPublic    string `json:"api_public"`
	Gateway      string `json:"gateway"`
	Media        string `json:"media"`
	UploadRelay  string `json:"upload_relay"`
	StaticCDN    string `json:"static_cdn"`
	Marketing    string `json:"marketing"`
	Admin        string `json:"admin"`
	Invite       string `json:"invite"`
	Gift         string `json:"gift"`
	Webapp       string `json:"webapp"`
}

// DiscoveryCaptcha details.
type DiscoveryCaptcha struct {
	Provider         string `json:"provider"`
	TurnstileSiteKey string `json:"turnstile_site_key,omitempty"`
}

// DiscoveryFeatures flags available in this instance.
type DiscoveryFeatures struct {
	VoiceEnabled                bool `json:"voice_enabled"`
	PremiumEnabled              bool `json:"premium_enabled"`
	SelfHosted                  bool `json:"self_hosted"`
	PresignedAttachmentUploads  bool `json:"presigned_attachment_uploads"`
	BoosterRoleEnabled          bool `json:"booster_role_enabled"`
}

// DiscoveryPush vapid configuration.
type DiscoveryPush struct {
	PublicVapidKey string `json:"public_vapid_key"`
}

// DiscoveryLegal links to terms and source.
type DiscoveryLegal struct {
	TermsURL   string `json:"terms_url"`
	PrivacyURL string `json:"privacy_url"`
	SourceURL  string `json:"source_url"`
}

// DiscoveryLimits holds rate limits and feature overrides.
type DiscoveryLimits struct {
	Version int `json:"version"`
}

// DiscoveryManager fetches and caches the discovery manifest.
type DiscoveryManager struct {
	httpClient *http.Client
	baseURL    string
	manifest   *DiscoveryManifest
	mu         sync.RWMutex
}

// NewDiscoveryManager creates a discovery manager for the given target base URL.
func NewDiscoveryManager(baseURL string, client *http.Client) *DiscoveryManager {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &DiscoveryManager{
		httpClient: client,
		baseURL:    baseURL,
	}
}

// Discover fetches the `.well-known/fluxer` endpoint.
func (d *DiscoveryManager) Discover(ctx context.Context) (*DiscoveryManifest, error) {
	d.mu.RLock()
	if d.manifest != nil {
		m := d.manifest
		d.mu.RUnlock()
		return m, nil
	}
	d.mu.RUnlock()

	d.mu.Lock()
	defer d.mu.Unlock()

	// Double check
	if d.manifest != nil {
		return d.manifest, nil
	}

	reqURL := d.baseURL + DefaultWellKnown
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery request: %w", err)
	}
	req.Header.Set("User-Agent", DefaultUserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("discovery request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discovery request returned status %d", resp.StatusCode)
	}

	var manifest DiscoveryManifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("failed to decode discovery JSON: %w", err)
	}

	d.manifest = &manifest
	return d.manifest, nil
}

// GetEndpoints returns the cached discovery endpoints, discovering them if necessary.
func (d *DiscoveryManager) GetEndpoints(ctx context.Context) (*DiscoveryEndpoints, error) {
	manifest, err := d.Discover(ctx)
	if err != nil {
		return nil, err
	}
	return &manifest.Endpoints, nil
}
