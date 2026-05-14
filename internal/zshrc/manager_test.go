package zshrc

import (
	"os"
	"strings"
	"testing"
)

func tmpZshrc(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "zshrc")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestWriteBlock_CreatesBlock(t *testing.T) {
	path := tmpZshrc(t, "export FOO=bar\n")
	if err := WriteBlock(path, map[string]string{"MyTools": "/usr/local/tools"}); err != nil {
		t.Fatal(err)
	}
	got := readFile(t, path)
	if !strings.Contains(got, blockBegin) || !strings.Contains(got, blockEnd) {
		t.Error("block markers not found")
	}
	if !strings.Contains(got, `export MyTools=`) {
		t.Error("export line not found")
	}
	// user content must be preserved
	if !strings.Contains(got, "export FOO=bar") {
		t.Error("existing user content was removed")
	}
}

func TestWriteBlock_Idempotent(t *testing.T) {
	path := tmpZshrc(t, "")
	exports := map[string]string{"A": "/a", "B": "/b"}

	if err := WriteBlock(path, exports); err != nil {
		t.Fatal(err)
	}
	first := readFile(t, path)

	if err := WriteBlock(path, exports); err != nil {
		t.Fatal(err)
	}
	second := readFile(t, path)

	if first != second {
		t.Errorf("second write changed the file:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
	if strings.Count(second, blockBegin) != 1 {
		t.Error("duplicate block markers found after second write")
	}
}

func TestWriteBlock_UpdatesExistingBlock(t *testing.T) {
	path := tmpZshrc(t, "# user stuff\n")
	if err := WriteBlock(path, map[string]string{"Old": "/old"}); err != nil {
		t.Fatal(err)
	}
	if err := WriteBlock(path, map[string]string{"New": "/new"}); err != nil {
		t.Fatal(err)
	}
	got := readFile(t, path)
	if strings.Contains(got, `export Old=`) {
		t.Error("old export should have been replaced")
	}
	if !strings.Contains(got, `export New=`) {
		t.Error("new export not found")
	}
	if !strings.Contains(got, "# user stuff") {
		t.Error("user content outside block was removed")
	}
}

func TestReadBlock_ReturnsLines(t *testing.T) {
	path := tmpZshrc(t, "")
	if err := WriteBlock(path, map[string]string{"X": "/x"}); err != nil {
		t.Fatal(err)
	}
	lines, err := ReadBlock(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) == 0 {
		t.Error("expected at least one line in block")
	}
}

func TestReadBlock_ReturnsNilWhenNoBlock(t *testing.T) {
	path := tmpZshrc(t, "export FOO=bar\n")
	lines, err := ReadBlock(path)
	if err != nil {
		t.Fatal(err)
	}
	if lines != nil {
		t.Errorf("expected nil, got %v", lines)
	}
}

func TestRemoveBlock(t *testing.T) {
	path := tmpZshrc(t, "# before\n")
	if err := WriteBlock(path, map[string]string{"X": "/x"}); err != nil {
		t.Fatal(err)
	}
	if err := RemoveBlock(path); err != nil {
		t.Fatal(err)
	}
	got := readFile(t, path)
	if strings.Contains(got, blockBegin) {
		t.Error("block marker still present after removal")
	}
	if !strings.Contains(got, "# before") {
		t.Error("content before block was lost")
	}
}

func TestWriteBlock_PreservesContentAfterBlock(t *testing.T) {
	initial := "# before\n" + blockBegin + "\nexport OLD=/old\n" + blockEnd + "\n# after\n"
	path := tmpZshrc(t, initial)
	if err := WriteBlock(path, map[string]string{"NEW": "/new"}); err != nil {
		t.Fatal(err)
	}
	got := readFile(t, path)
	if !strings.Contains(got, "# before") {
		t.Error("content before block lost")
	}
	if !strings.Contains(got, "# after") {
		t.Error("content after block lost")
	}
}
