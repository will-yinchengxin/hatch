package fsutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadTextFile(t *testing.T) {
	dir := t.TempDir()
	write := func(name string, data []byte) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, data, 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}

	tf, err := ReadTextFile(write("ok.sh", []byte("#!/bin/zsh\necho 你好\n")), 1024)
	if err != nil || tf.Binary || tf.Truncated || tf.Content != "#!/bin/zsh\necho 你好\n" {
		t.Fatalf("text file: %+v %v", tf, err)
	}

	tf, _ = ReadTextFile(write("bin", []byte("\xcf\xfa\xed\xfe\x00\x00\x01")), 1024)
	if !tf.Binary || tf.Content != "" || tf.Size != 7 {
		t.Fatalf("binary file: %+v", tf)
	}

	// "你" is 3 bytes; cutting at 4 bytes must not split it into invalid UTF-8.
	tf, _ = ReadTextFile(write("big.txt", []byte("a你好"+strings.Repeat("x", 100))), 4)
	if tf.Binary || !tf.Truncated || tf.Content != "a你" {
		t.Fatalf("truncated file: %+v", tf)
	}
}
