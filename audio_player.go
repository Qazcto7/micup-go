package micup

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"
)

const (
	// Standard Micup / LiveKit audio parameters
	DefaultSampleRate    = 48000
	DefaultChannels      = 2
	FrameDurationMs      = 20
	SamplesPerFrame      = (DefaultSampleRate * FrameDurationMs) / 1000 // 960 samples per channel
	BytesPerSample       = 2                                            // 16-bit signed PCM
	FrameByteSize        = SamplesPerFrame * DefaultChannels * BytesPerSample // 3840 bytes
)

// AudioPlayer manages high-level audio playback, FFmpeg transcoding, and volume control.
type AudioPlayer struct {
	vc        *VoiceConnection
	mu        sync.RWMutex
	playing   bool
	paused    bool
	volume    float64
	cmd       *exec.Cmd
	stopChan  chan struct{}
	pauseChan chan struct{}
	onFinish  func()
}

// NewAudioPlayer creates a new audio player bound to a VoiceConnection.
func NewAudioPlayer(vc *VoiceConnection) *AudioPlayer {
	return &AudioPlayer{
		vc:        vc,
		volume:    1.0,
		stopChan:  make(chan struct{}),
		pauseChan: make(chan struct{}),
	}
}

// SetVolume adjusts playback volume. 1.0 = 100%, 0.5 = 50%, 2.0 = 200%.
func (p *AudioPlayer) SetVolume(volume float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if volume < 0 {
		volume = 0
	}
	p.volume = volume
}

// Volume returns current volume.
func (p *AudioPlayer) Volume() float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.volume
}

// IsPlaying returns whether audio is currently streaming.
func (p *AudioPlayer) IsPlaying() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.playing && !p.paused
}

// IsPaused returns whether audio playback is currently paused.
func (p *AudioPlayer) IsPaused() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.paused
}

// Pause pauses active audio playback.
func (p *AudioPlayer) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.playing && !p.paused {
		p.paused = true
	}
}

// Resume resumes paused playback.
func (p *AudioPlayer) Resume() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.playing && p.paused {
		p.paused = false
	}
}

// Stop terminates active playback immediately.
func (p *AudioPlayer) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stopLocked()
}

func (p *AudioPlayer) stopLocked() {
	if !p.playing {
		return
	}
	p.playing = false
	p.paused = false

	if p.stopChan != nil {
		select {
		case <-p.stopChan:
		default:
			close(p.stopChan)
		}
	}

	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
		p.cmd = nil
	}

	_ = p.vc.SetSpeaking(false)

	if p.onFinish != nil {
		go p.onFinish()
	}
}

// OnFinish sets a callback triggered when playback ends.
func (p *AudioPlayer) OnFinish(handler func()) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.onFinish = handler
}

// Play launches FFmpeg to stream any audio file or remote URL to the voice channel.
func (p *AudioPlayer) Play(source string) error {
	return p.PlayContext(context.Background(), source)
}

// PlayContext launches playback with context cancellation.
func (p *AudioPlayer) PlayContext(ctx context.Context, source string) error {
	p.mu.Lock()
	if p.playing {
		p.stopLocked()
	}

	p.playing = true
	p.paused = false
	p.stopChan = make(chan struct{})

	// ffmpeg command: transcode input to raw 48kHz 16-bit stereo PCM
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", source,
		"-f", "s16le",
		"-ar", fmt.Sprintf("%d", DefaultSampleRate),
		"-ac", fmt.Sprintf("%d", DefaultChannels),
		"-loglevel", "warning",
		"pipe:1",
	)
	p.cmd = cmd
	p.mu.Unlock()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		p.Stop()
		return fmt.Errorf("failed to open ffmpeg stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		p.Stop()
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	_ = p.vc.SetSpeaking(true)

	go p.streamLoop(stdout)
	return nil
}

// streamLoop reads raw PCM from ffmpeg and queues it at 20ms intervals.
func (p *AudioPlayer) streamLoop(reader io.Reader) {
	defer func() {
		p.Stop()
	}()

	ticker := time.NewTicker(FrameDurationMs * time.Millisecond)
	defer ticker.Stop()

	frameBuf := make([]byte, FrameByteSize)

	for {
		select {
		case <-p.stopChan:
			return
		case <-ticker.C:
			p.mu.RLock()
			isPaused := p.paused
			vol := p.volume
			p.mu.RUnlock()

			if isPaused {
				continue
			}

			n, err := io.ReadFull(reader, frameBuf)
			if err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					return
				}
				return
			}

			// Apply volume scaling if not 100%
			payload := frameBuf[:n]
			if vol != 1.0 {
				payload = ScalePCM16(payload, vol)
			}

			// Queue frame to VoiceConnection
			_ = p.vc.SendOpusFrame(payload)
		}
	}
}

// ScalePCM16 scales 16-bit signed PCM audio samples by a volume multiplier with clamping.
func ScalePCM16(pcm []byte, volume float64) []byte {
	if volume == 1.0 || len(pcm) < 2 {
		return pcm
	}

	out := make([]byte, len(pcm))
	for i := 0; i < len(pcm)-1; i += 2 {
		sample := int16(binary.LittleEndian.Uint16(pcm[i : i+2]))
		scaled := float64(sample) * volume

		// Clamp to int16 boundaries
		if scaled > 32767 {
			scaled = 32767
		} else if scaled < -32768 {
			scaled = -32768
		}

		binary.LittleEndian.PutUint16(out[i:i+2], uint16(int16(scaled)))
	}
	return out
}

// NewPlayer adds a helper to create an AudioPlayer directly from VoiceConnection.
func (vc *VoiceConnection) NewPlayer() *AudioPlayer {
	return NewAudioPlayer(vc)
}
