package hostmonitor

import (
	"os"
	"testing"
)

func TestDiskBenchmarkCleansTemporaryFile(t *testing.T) {
	dir := t.TempDir()
	result, err := benchmarkDisk(dir, 1024*1024)
	if err != nil {
		t.Fatal(err)
	}
	if result.Bytes != 1024*1024 || result.WriteBytesPerSecond <= 0 || result.ReadBytesPerSecond <= 0 {
		t.Fatalf("unexpected result: %#v", result)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		t.Fatalf("temporary files remain: %v %v", entries, err)
	}
}

func TestDiskBenchmarkRejectsInvalidSize(t *testing.T) {
	if _, err := benchmarkDisk(t.TempDir(), 0); err == nil {
		t.Fatal("expected error")
	}
}
