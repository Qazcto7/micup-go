package micup

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"
)

// AudioFrame represents an Opus-encoded audio frame.
type AudioFrame struct {
	Opus []byte
}

// AudioStreamer represents an audio transmitter to a Micup voice channel.
type AudioStreamer struct {
	voiceConn *VoiceConnection
	frames    chan []byte
	stopChan  chan struct{}
	mu        sync.Mutex
	streaming bool
}

// NewAudioStreamer initializes an audio streamer for a voice connection.
func NewAudioStreamer(vc *VoiceConnection) *AudioStreamer {
	return &AudioStreamer{
		voiceConn: vc,
		frames:    make(chan []byte, 100),
		stopChan:  make(chan struct{}),
	}
}

// SendOpusFrame queues an Opus-encoded audio frame (typically 20ms of audio).
func (a *AudioStreamer) SendOpusFrame(frame []byte) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	select {
	case <-a.stopChan:
		return errors.New("audio streamer is closed")
	default:
	}

	select {
	case a.frames <- frame:
		return nil
	default:
		return errors.New("audio buffer full, frame dropped")
	}
}

// StreamPCMOrOpus streams audio frames from an io.Reader at regular intervals (20ms default).
func (a *AudioStreamer) StreamOpusReader(ctx context.Context, reader io.Reader, frameSize int, interval time.Duration) error {
	if interval <= 0 {
		interval = 20 * time.Millisecond // Standard Opus 20ms frame
	}
	if frameSize <= 0 {
		frameSize = 960 // Typical 48kHz stereo/mono buffer
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	buf := make([]byte, frameSize)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-a.stopChan:
			return nil
		case <-ticker.C:
			n, err := io.ReadFull(reader, buf)
			if err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					return nil
				}
				return err
			}
			_ = a.SendOpusFrame(buf[:n])
		}
	}
}

// Close stops the audio stream.
func (a *AudioStreamer) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	select {
	case <-a.stopChan:
		return nil
	default:
		close(a.stopChan)
	}
	return nil
}
