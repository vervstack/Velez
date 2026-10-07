package address_book

import (
	"sync"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
)

var errRebuildPanicked = rerrors.New("addresses rebuild panicked")

// rebuildProgress is both the single-flight guard and the progress tracker of
// a rebuild. Every method is safe on a nil receiver, so callers that don't
// report progress (Sync) pass nil.
type rebuildProgress struct {
	mu     sync.Mutex
	status domain.AddressRebuildStatus
}

func newRebuildProgress() *rebuildProgress {
	return &rebuildProgress{}
}

func (p *rebuildProgress) start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status.IsRunning {
		return rerrors.Wrap(user_errors.ErrAddressesRebuildInProgress)
	}

	p.status = domain.AddressRebuildStatus{IsRunning: true}

	return nil
}

func (p *rebuildProgress) addSteps(count int) {
	if p == nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.status.TotalSteps += uint32(count)
}

func (p *rebuildProgress) step() {
	if p == nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status.DoneSteps < p.status.TotalSteps {
		p.status.DoneSteps++
	}
}

func (p *rebuildProgress) finish(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.status.IsRunning = false

	if err != nil {
		p.status.LastError = err.Error()

		return
	}

	p.status.LastError = ""
	p.status.DoneSteps = p.status.TotalSteps
}

func (p *rebuildProgress) snapshot() domain.AddressRebuildStatus {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.status
}
