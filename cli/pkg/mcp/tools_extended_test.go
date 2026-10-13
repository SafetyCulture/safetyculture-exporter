package mcp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadJailedFile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "report.pdf"), []byte("pdf-bytes"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sub"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sub", "photo.png"), []byte("png-bytes"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "secret.yaml"), []byte("token: abc"), 0644))

	outside := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outside, "evil.pdf"), []byte("nope"), 0644))

	// OK: top-level allowed ext
	data, name, mime, err := readJailedFile(dir, "report.pdf", []string{".pdf"})
	require.NoError(t, err)
	assert.Equal(t, "pdf-bytes", string(data))
	assert.Equal(t, "report.pdf", name)
	assert.Equal(t, "application/pdf", mime)

	// OK: nested allowed ext
	_, _, mime, err = readJailedFile(dir, "sub/photo.png", []string{".png"})
	require.NoError(t, err)
	assert.Equal(t, "image/png", mime)

	// Reject: traversal
	_, _, _, err = readJailedFile(dir, "../secret.pdf", []string{".pdf"})
	assert.ErrorContains(t, err, "escapes")

	// Reject: absolute path
	_, _, _, err = readJailedFile(dir, filepath.Join(outside, "evil.pdf"), []string{".pdf"})
	assert.Error(t, err)

	// Reject: disallowed extension
	_, _, _, err = readJailedFile(dir, "secret.yaml", []string{".pdf"})
	assert.ErrorContains(t, err, "not allowed")

	// Reject: symlink escape
	link := filepath.Join(dir, "linked.pdf")
	require.NoError(t, os.Symlink(filepath.Join(outside, "evil.pdf"), link))
	_, _, _, err = readJailedFile(dir, "linked.pdf", []string{".pdf"})
	assert.ErrorContains(t, err, "escapes")

	// Reject: missing file
	_, _, _, err = readJailedFile(dir, "nope.pdf", []string{".pdf"})
	assert.Error(t, err)
}

func TestMimeForExt(t *testing.T) {
	assert.Equal(t, "application/pdf", mimeForExt(".pdf"))
	assert.Equal(t, "image/jpeg", mimeForExt(".jpg"))
	assert.Equal(t, "application/octet-stream", mimeForExt(".bin"))
}

func TestListReportFiles(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.pdf"), []byte("x"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.docx"), []byte("x"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ignored.txt"), []byte("x"), 0644))

	files, err := ListReportFiles(dir)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"a.docx", "b.pdf"}, files)
}
