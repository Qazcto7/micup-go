package micup

import (
	"os"
	"os/exec"
	"strings"
)

// HardwareEncoderCapability holds probed hardware acceleration features.
type HardwareEncoderCapability struct {
	HasNVENC        bool     `json:"has_nvenc"`
	HasVAAPI        bool     `json:"has_vaapi"`
	SupportedCodecs []string `json:"supported_codecs"`
	PreferredCodec  string   `json:"preferred_codec"`
	DriverInfo      string   `json:"driver_info,omitempty"`
}

// ProbeHardwareEncoders checks for available hardware video encoders on the host system
// (matching @fluxer/hardware-encoder reverse engineering).
func ProbeHardwareEncoders() *HardwareEncoderCapability {
	cap := &HardwareEncoderCapability{
		SupportedCodecs: make([]string, 0),
		PreferredCodec:  CodecH264,
	}

	// 1. Probe NVIDIA NVENC
	nvencLibs := []string{
		"/usr/lib/libnvidia-encode.so.1",
		"/usr/lib64/libnvidia-encode.so.1",
		"/usr/lib/x86_64-linux-gnu/libnvidia-encode.so.1",
	}
	for _, lib := range nvencLibs {
		if _, err := os.Stat(lib); err == nil {
			cap.HasNVENC = true
			break
		}
	}
	if !cap.HasNVENC {
		if _, err := os.Stat("/dev/nvidiactl"); err == nil {
			cap.HasNVENC = true
		}
	}

	// 2. Probe VA-API / DRM
	if _, err := os.Stat("/dev/dri/renderD128"); err == nil {
		cap.HasVAAPI = true
	}

	// 3. Query ffmpeg supported encoders if installed
	if out, err := exec.Command("ffmpeg", "-encoders").Output(); err == nil {
		outStr := string(out)
		if strings.Contains(outStr, "h264_nvenc") || strings.Contains(outStr, "h264_vaapi") || strings.Contains(outStr, "libx264") {
			cap.SupportedCodecs = append(cap.SupportedCodecs, CodecH264)
		}
		if strings.Contains(outStr, "hevc_nvenc") || strings.Contains(outStr, "hevc_vaapi") || strings.Contains(outStr, "libx265") {
			cap.SupportedCodecs = append(cap.SupportedCodecs, "hevc")
		}
		if strings.Contains(outStr, "libvpx") {
			cap.SupportedCodecs = append(cap.SupportedCodecs, CodecVP8, CodecVP9)
		}
		if strings.Contains(outStr, "av1_nvenc") || strings.Contains(outStr, "libsvtav1") || strings.Contains(outStr, "av1_vaapi") {
			cap.SupportedCodecs = append(cap.SupportedCodecs, CodecAV1)
		}
	} else {
		// Fallback baseline
		cap.SupportedCodecs = append(cap.SupportedCodecs, CodecH264, CodecVP8)
	}

	if len(cap.SupportedCodecs) > 0 {
		cap.PreferredCodec = cap.SupportedCodecs[0]
	}

	return cap
}
