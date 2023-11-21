package tarball_test

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/abcdlsj/kiwi/internal/tarball"
)

func TestTarDir(t *testing.T) {
	ctx := context.Background()

	buf, err := tarball.TarDir(ctx, "./")
	if err != nil {
		t.Fatal(err)
	}

	archiveTmp, _ := os.CreateTemp("", "*.tar")
	t.Logf("archive: %s", archiveTmp.Name())

	io.Copy(archiveTmp, buf)
}
