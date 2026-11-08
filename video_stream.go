package micup

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"
)

// VideoFrame represents an encoded video packet (e.g. H.264 NAL unit or VP8 frame).
type VideoFrame struct {
	Data      []byte
	IsKey     bool
	Width     int
	Height    int
	Timestamp time.Duration
}

// VideoStreamer manages real-time video frame queueing and FFmpeg transcoding to voice channels.
type VideoStreamer struct {
	voiceConn *VoiceConnection
	frames    chan []byte
	stopChan  chan struct{}
	mu        sync.Mutex
	streaming bool
	Width     int
	Height    int
	FPS       int
	Bitrate   int
	cmd       *exec.Cmd
}

// NewVideoStreamer initializes a video streamer for a voice connection.
func NewVideoStreamer(vc *VoiceConnection, width, height, fps, bitrateKbps int) *VideoStreamer {
	if width <= 0 {
		width = 1280
	}
	if height <= 0 {
		height = 720
	}
	if fps <= 0 {
		fps = 30
	}
	if bitrateKbps <= 0 {
		bitrateKbps = 3000
	}

	return &VideoStreamer{
		voiceConn: vc,
		frames:    make(chan []byte, 120),
		stopChan:  make(chan struct{}),
		Width:     width,
		Height:    height,
		FPS:       fps,
		Bitrate:   bitrateKbps,
	}
}

// SendVideoFrame queues an encoded video packet for transmission.
func (v *VideoStreamer) SendVideoFrame(frame []byte) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	select {
	case <-v.stopChan:
		return errors.New("video streamer is closed")
	default:
	}

	select {
	case v.frames <- frame:
		return nil
	default:
		// Drop frame on backpressure to keep real-time latency minimal
		return errors.New("video frame queue full, frame dropped")
	}
}

// StreamFile transcodes any local video file or remote stream into H.264 NAL units and streams it at the configured FPS.
func (v *VideoStreamer) StreamFile(ctx context.Context, source string) error {
	v.mu.Lock()
	if v.streaming {
		v.mu.Unlock()
		return errors.New("video streaming already in progress")
	}
	v.streaming = true
	v.stopChan = make(chan struct{})

	// FFmpeg transcode: baseline H.264 Annex B stream into stdout pipe
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-re", // Read input at native frame rate
		"-i", source,
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-preset", "ultrafast",
		"-tune", "zerolatency",
		"-b:v", fmt.Sprintf("%dk", v.Bitrate),
		"-maxrate", fmt.Sprintf("%dk", v.Bitrate),
		"-bufsize", fmt.Sprintf("%dk", v.Bitrate*2),
		"-g", fmt.Sprintf("%d", v.FPS*2), // Keyframe interval
		"-r", fmt.Sprintf("%d", v.FPS),
		"-s", fmt.Sprintf("%dx%d", v.Width, v.Height),
		"-f", "h264",
		"-loglevel", "warning",
		"pipe:1",
	)
	v.cmd = cmd
	v.mu.Unlock()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = v.Close()
		return fmt.Errorf("failed to open ffmpeg video stdout: %w", err)
	}

	if err := cmd.Start(); err != nil {
		_ = v.Close()
		return fmt.Errorf("failed to start ffmpeg video transcoder: %w", err)
	}

	go v.readNALUnits(stdout)
	return nil
}

// readNALUnits splits raw H.264 Annex-B stream by 0x00000001 start codes.
func (v *VideoStreamer) readNALUnits(reader io.Reader) {
	defer func() {
		_ = v.Close()
	}()

	interval := time.Second / time.Duration(v.FPS)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	br := bufio.NewReaderSize(reader, 256*1024)
	startCode := []byte{0x00, 0x00, 0x00, 0x01}

	var nalBuffer bytes.Buffer

	for {
		select {
		case <-v.stopChan:
			return
		case <-ticker.C:
			// Read chunk from FFmpeg
			chunk := make([]byte, 16384)
			n, err := br.Read(chunk)
			if err != nil {
				return
			}

			nalBuffer.Write(chunk[:n])

			// Extract complete NAL packets
			for {
				data := nalBuffer.Bytes()
				idx := bytes.Index(data[4:], startCode)
				if idx == -1 {
					break
				}
				packet := make([]byte, idx+4)
				copy(packet, data[:idx+4])

				_ = v.SendVideoFrame(packet)
				nalBuffer.Next(idx + 4)
			}
		}
	}
}

// Close stops the video stream and kills transcoding processes.
func (v *VideoStreamer) Close() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if !v.streaming {
		return nil
	}
	v.streaming = false

	select {
	case <-v.stopChan:
	default:
		close(v.stopChan)
	}

	if v.cmd != nil && v.cmd.Process != nil {
		_ = v.cmd.Process.Kill()
		v.cmd = nil
	}
	return nil
}
