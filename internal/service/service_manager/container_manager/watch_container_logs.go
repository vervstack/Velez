package container_manager

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/docker/docker/pkg/stdcopy"
	"go.redsock.ru/rerrors"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/utils/common"
)

const (
	defaultLogsTail = 100
)

// WatchContainerLogs streams the container's logs by literal Docker id, with
// the same unscoped lookup as GetContainer. It returns nil when the Docker
// stream ends or ctx is cancelled, and send's error when send fails.
func (c *ContainerManager) WatchContainerLogs(
	ctx context.Context,
	req *velez_api.WatchContainerLogs_Request,
	send func(*velez_api.WatchContainerLogs_Response) error,
) error {
	runtime, err := c.runtimes.Runtime(ctx, req.GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving environment")
	}

	contInfo, found, err := runtime.InspectAny(ctx, req.GetId())
	if err != nil {
		return rerrors.Wrap(err, "error inspecting container")
	}

	if !found {
		return rerrors.Wrap(user_errors.ErrNoSuchContainer)
	}

	tail := req.GetTail()
	if tail == 0 {
		tail = defaultLogsTail
	}

	logsReq := container_runtime.LogsRequest{
		Tail:     tail,
		IsFollow: req.GetIsFollow(),
	}

	logs, err := runtime.Logs(ctx, contInfo.ID, logsReq)
	if err != nil {
		return rerrors.Wrap(err, "error opening container logs")
	}
	defer common.CloseWithLog(logs.Close, "container logs stream")

	stdout := &logLineWriter{stream: velez_api.LogStream_STDOUT, send: send}
	stderr := &logLineWriter{stream: velez_api.LogStream_STDERR, send: send}

	// A TTY container's stream is raw; otherwise Docker frames stdout/stderr (stdcopy).
	if contInfo.Config != nil && contInfo.Config.Tty {
		_, err = io.Copy(stdout, logs)
	} else {
		_, err = stdcopy.StdCopy(stdout, stderr, logs)
	}

	if ctx.Err() != nil {
		return nil
	}

	if err != nil {
		return rerrors.Wrap(err, "error reading container logs")
	}

	err = stdout.flush()
	if err != nil {
		return rerrors.Wrap(err, "error sending stdout tail")
	}

	err = stderr.flush()
	if err != nil {
		return rerrors.Wrap(err, "error sending stderr tail")
	}

	return nil
}

// logLineWriter splits a byte stream into lines and sends each as a response.
// With Timestamps on, Docker prefixes every line with an RFC3339Nano timestamp
// and a single space.
type logLineWriter struct {
	stream velez_api.LogStream
	send   func(*velez_api.WatchContainerLogs_Response) error
	buf    []byte
}

func (w *logLineWriter) Write(p []byte) (int, error) {
	w.buf = append(w.buf, p...)

	for {
		idx := bytes.IndexByte(w.buf, '\n')
		if idx < 0 {
			return len(p), nil
		}

		line := w.buf[:idx]

		w.buf = w.buf[idx+1:]

		err := w.sendLine(line)
		if err != nil {
			return 0, rerrors.Wrap(err, "error sending log line")
		}
	}
}

func (w *logLineWriter) flush() error {
	if len(w.buf) == 0 {
		return nil
	}

	line := w.buf

	w.buf = nil

	return w.sendLine(line)
}

func (w *logLineWriter) sendLine(line []byte) error {
	line = bytes.TrimSuffix(line, []byte("\r"))

	resp := &velez_api.WatchContainerLogs_Response{
		Stream: w.stream,
		Text:   string(line),
	}

	rawTimestamp, text, hasTimestamp := bytes.Cut(line, []byte(" "))
	if hasTimestamp {
		parsed, err := time.Parse(time.RFC3339Nano, string(rawTimestamp))
		if err == nil {
			resp.Timestamp = timestamppb.New(parsed)
			resp.Text = string(text)
		}
	}

	return w.send(resp)
}
