package hostmonitor

import "testing"

func TestParseNetworkQuality(t *testing.T) {
	result, err := parseNetworkQuality([]byte(`{"dl_throughput":100000000,"ul_throughput":20000000,"base_rtt":24.5,"dl_bytes_transferred":1024,"ul_bytes_transferred":512,"interface_name":"en0"}`))
	if err != nil || result.DownloadBPS != 100000000 || result.UploadBPS != 20000000 || result.LatencyMS != 24.5 || result.DownloadBytes != 1024 || result.Interface != "en0" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if _, err := parseNetworkQuality([]byte(`{"base_rtt":2}`)); err == nil {
		t.Fatal("missing throughput must fail")
	}
}
