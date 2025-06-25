package dockerutils

import (
	"archive/tar"
	"bytes"
	"context"
	stderrs "errors"
	"io"
	"path"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"
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
			stderrs.Join(err, errClose)
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

func WriteToContainer(ctx context.Context, dockerAPI client.APIClient, contId string, systemPath string, content []byte) error {

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	hdr := &tar.Header{
		Name:    path.Base(systemPath),
		Mode:    0644,
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
