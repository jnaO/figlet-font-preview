package newfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSafeName(t *testing.T) {
	if got := SafeName(`C64-fonts/gothic:a*b?"<>|\\x`); got != "C64-fonts_gothic_a_b_______x" {
		t.Fatalf("SafeName = %q", got)
	}
}

func TestWriteNeverOverwrites(t *testing.T) {
	dir := t.TempDir()
	first, err := Write(dir, "x", ".svg", []byte("one"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := Write(dir, "x", ".svg", []byte("two"))
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, "x.svg"); first != want || !filepath.IsAbs(first) {
		t.Fatalf("first = %q, want %q", first, want)
	}
	if want := filepath.Join(dir, "x-2.svg"); second != want || !filepath.IsAbs(second) {
		t.Fatalf("second = %q, want %q", second, want)
	}
	if b, _ := os.ReadFile(first); string(b) != "one" {
		t.Fatalf("first file now %q", b)
	}
	if b, _ := os.ReadFile(second); string(b) != "two" {
		t.Fatalf("second file %q", b)
	}
}
