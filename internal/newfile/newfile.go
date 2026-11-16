// Package newfile writes files without ever overwriting an existing one.
package newfile

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// SafeName replaces characters that are unsafe in file names with '_'.
func SafeName(name string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 || strings.ContainsRune(`/\:*?"<>|`, r) {
			return '_'
		}
		return r
	}, name)
}

// Write creates dir/base+ext, or base-2, base-3, … when taken, and
// returns its absolute path. O_EXCL makes the existence check atomic.
func Write(dir, base, ext string, data []byte) (string, error) {
	for n := 1; ; n++ {
		name := base + ext
		if n > 1 {
			name = fmt.Sprintf("%s-%d%s", base, n, ext)
		}
		path := filepath.Join(dir, name)
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if errors.Is(err, fs.ErrExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		_, werr := f.Write(data)
		if cerr := f.Close(); werr == nil {
			werr = cerr
		}
		if werr != nil {
			return "", werr
		}
		return filepath.Abs(path)
	}
}
