//go:generate minimock -i Worker -o ./../../tests/mocks -g -s "_mock.go"

package workers

import (
	"context"
)

type Worker interface {
	Start(ctx context.Context)
	Stop() error
}
