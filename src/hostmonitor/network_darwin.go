package hostmonitor

import (
	"encoding/json"
	"fmt"
	"time"
)

type NetworkQuality struct {
	DownloadBPS   float64 `json:"downloadBps"`
	UploadBPS     float64 `json:"uploadBps"`
	LatencyMS     float64 `json:"latencyMs"`
	DownloadBytes uint64  `json:"downloadBytes"`
	UploadBytes   uint64  `json:"uploadBytes"`
	Interface     string  `json:"interface"`
}

func parseNetworkQuality(data []byte) (NetworkQuality, error) {
	var raw struct {
		Download      *float64 `json:"dl_throughput"`
		Upload        *float64 `json:"ul_throughput"`
		Latency       float64  `json:"base_rtt"`
		DownloadBytes uint64   `json:"dl_bytes_transferred"`
		UploadBytes   uint64   `json:"ul_bytes_transferred"`
		Interface     string   `json:"interface_name"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return NetworkQuality{}, fmt.Errorf("无法解析网络测速结果: %w", err)
	}
	if raw.Download == nil || raw.Upload == nil || *raw.Download < 0 || *raw.Upload < 0 {
		return NetworkQuality{}, fmt.Errorf("测速结果缺少上传或下载速率")
	}
	return NetworkQuality{DownloadBPS: *raw.Download, UploadBPS: *raw.Upload, LatencyMS: raw.Latency, DownloadBytes: raw.DownloadBytes, UploadBytes: raw.UploadBytes, Interface: raw.Interface}, nil
}

func MeasureNetwork() (NetworkQuality, error) {
	data, err := command(120*time.Second, "/usr/bin/networkQuality", "-s", "-c")
	if err != nil {
		return NetworkQuality{}, fmt.Errorf("系统网络测速失败或超时: %w", err)
	}
	return parseNetworkQuality(data)
}
