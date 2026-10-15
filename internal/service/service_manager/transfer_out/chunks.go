package transfer_out

import (
	"errors"
	"io"

	"go.redsock.ru/rerrors"
)

const (
	chunkSize = 1 << 20
)

// forEachChunk cuts the reader into chunks of chunkBytes (the last one may be shorter). The chunk
// slice is reused between calls, so emit must not retain it.
func forEachChunk(reader io.Reader, chunkBytes int, emit func(chunk []byte) error) error {
	buf := make([]byte, chunkBytes)

	for {
		n, err := io.ReadFull(reader, buf)
		if n > 0 {
			emitErr := emit(buf[:n])
			if emitErr != nil {
				return rerrors.Wrap(emitErr)
			}
		}

		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil
		}

		if err != nil {
			return rerrors.Wrap(err, "error reading item")
		}
	}
}
