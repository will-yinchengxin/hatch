package hostmonitor

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"time"
)

type DiskBenchmark struct {
	Bytes               int64   `json:"bytes"`
	WriteBytesPerSecond float64 `json:"writeBytesPerSecond"`
	ReadBytesPerSecond  float64 `json:"readBytesPerSecond"`
}

func BenchmarkDisk() (DiskBenchmark, error) {
	return benchmarkDisk(os.TempDir(), 64*1024*1024)
}

func benchmarkDisk(dir string, size int64) (DiskBenchmark, error) {
	if size <= 0 || size > 64*1024*1024 {
		return DiskBenchmark{}, fmt.Errorf("invalid benchmark size")
	}
	file, err := os.CreateTemp(dir, "hatch-disk-test-*")
	if err != nil {
		return DiskBenchmark{}, err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	buffer := make([]byte, 1024*1024)
	if _, err = rand.Read(buffer); err != nil {
		return DiskBenchmark{}, err
	}
	start := time.Now()
	for remaining := size; remaining > 0; {
		chunk := int64(len(buffer))
		if remaining < chunk {
			chunk = remaining
		}
		written, writeErr := file.Write(buffer[:chunk])
		if writeErr != nil {
			return DiskBenchmark{}, writeErr
		}
		if int64(written) != chunk {
			return DiskBenchmark{}, io.ErrShortWrite
		}
		remaining -= int64(written)
	}
	if err = file.Sync(); err != nil {
		return DiskBenchmark{}, err
	}
	writeSeconds := time.Since(start).Seconds()
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return DiskBenchmark{}, err
	}
	start = time.Now()
	read, err := io.CopyBuffer(io.Discard, file, buffer)
	if err != nil {
		return DiskBenchmark{}, err
	}
	if read != size {
		return DiskBenchmark{}, io.ErrUnexpectedEOF
	}
	readSeconds := time.Since(start).Seconds()
	return DiskBenchmark{Bytes: size, WriteBytesPerSecond: float64(size) / writeSeconds, ReadBytesPerSecond: float64(size) / readSeconds}, nil
}
