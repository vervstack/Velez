package dockerutils

import (
	"archive/tar"
	"bytes"
	"context"
	stderrs "errors"
	"io"
	"path"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	tarFilePermissions = 0o644

	// maxDirReadBytes caps the total decompressed content ReadDirFromContainer
	// buffers in memory, guarding against a pathological or malicious tar
	// stream claiming an unbounded amount of data.
	maxDirReadBytes = 32 * 1024 * 1024
)

func ReadFromContainer(ctx context.Context, dockerAPI client.APIClient, contId string, path string) ([]byte, error) {
	rc, _, err := dockerAPI.CopyFromContainer(ctx, contId, path)
	if err != nil {
		return nil, rerrors.Wrap(err, "error coping from container")
	}

	defer func() {
		errClose := rc.Close()
		if errClose == nil {
			return
		}

		if err == nil {
			err = errClose
		} else {
			err = stderrs.Join(err, errClose)
		}
	}()

	reader := tar.NewReader(rc)

	_, err = reader.Next()
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting next")
	}

	res, err := io.ReadAll(reader)
	if err != nil {
		return nil, rerrors.Wrap(err, "error reading config from tar")
	}

	return res, nil
}

// ReadDirFromContainer copies dirPath out of the container and returns every
// regular file under it, keyed by its path relative to dirPath (forward
// slashes, no leading slash). Docker's tar stream prefixes every entry with
// the basename of the copied path, e.g. copying "/verv" yields entries like
// "verv/vervonomicon.yaml" and "verv/prod/ingress.conf" - that leading
// component is stripped so callers get "vervonomicon.yaml" and
// "prod/ingress.conf".
//
// Directory entries, symlinks, and any other non-regular tar entry are
// skipped. An entry whose relative path escapes dirPath (contains "..") is
// dropped rather than written into the result. The total decompressed size is
// capped at maxDirReadBytes to guard against a pathological archive.
//
// If dirPath does not exist in the container, the returned error keeps the
// classification CopyFromContainer reports - callers can still check it with
// errdefs.IsNotFound, since rerrors.Wrap preserves the Unwrap chain.
func ReadDirFromContainer(
	ctx context.Context,
	dockerAPI client.APIClient,
	contId string,
	dirPath string,
) (map[string][]byte, error) {
	rc, _, err := dockerAPI.CopyFromContainer(ctx, contId, dirPath)
	if err != nil {
		return nil, rerrors.Wrap(err, "error copying directory from container")
	}

	defer func() {
		errClose := rc.Close()
		if errClose == nil {
			return
		}

		if err == nil {
			err = errClose
		} else {
			err = stderrs.Join(err, errClose)
		}
	}()

	result := make(map[string][]byte)
	reader := tar.NewReader(rc)
	totalBytes := int64(0)

	for {
		var hdr *tar.Header

		hdr, err = reader.Next()
		if stderrs.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, rerrors.Wrap(err, "error reading next tar entry")
		}

		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		separatorIdx := strings.Index(hdr.Name, "/")
		if separatorIdx < 0 {
			continue
		}

		relPath := path.Clean(hdr.Name[separatorIdx+1:])
		if relPath == ".." || strings.HasPrefix(relPath, "../") {
			continue
		}

		totalBytes += hdr.Size
		if totalBytes > maxDirReadBytes {
			return nil, user_errors.ErrDirReadSizeExceeded
		}

		var content []byte

		content, err = io.ReadAll(reader)
		if err != nil {
			return nil, rerrors.Wrap(err, "error reading tar entry content")
		}

		result[relPath] = content
	}

	return result, nil
}

func WriteToContainer(
	ctx context.Context,
	dockerAPI client.APIClient,
	contId string,
	systemPath string,
	content []byte,
) error {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	hdr := &tar.Header{
		Name:    path.Base(systemPath),
		Mode:    tarFilePermissions,
		Size:    int64(len(content)),
		ModTime: time.Now(),
	}

	err := tw.WriteHeader(hdr)
	if err != nil {
		return rerrors.Wrap(err, "error writing tar header")
	}

	_, err = tw.Write(content)
	if err != nil {
		return rerrors.Wrap(err, "error writing content")
	}

	err = tw.Close()
	if err != nil {
		return rerrors.Wrap(err, "error closing tar writer")
	}

	err = dockerAPI.CopyToContainer(ctx,
		contId, path.Dir(systemPath), buf,
		container.CopyToContainerOptions{})
	if err != nil {
		return rerrors.Wrap(err, "error writing to container")
	}

	return nil
}
