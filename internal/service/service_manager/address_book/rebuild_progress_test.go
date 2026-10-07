package address_book

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/user_errors"
)

var errTestRebuildFailed = rerrors.New("test rebuild failed")

func Test_RebuildProgress_SecondStartIsRejected(t *testing.T) {
	progress := newRebuildProgress()

	err := progress.start()
	require.NoError(t, err)

	err = progress.start()
	require.ErrorIs(t, err, user_errors.ErrAddressesRebuildInProgress)
}

func Test_RebuildProgress_GuardIsReleasedAfterFinish(t *testing.T) {
	progress := newRebuildProgress()

	require.NoError(t, progress.start())
	progress.finish(nil)

	require.NoError(t, progress.start())
}

func Test_RebuildProgress_GuardIsReleasedAfterError(t *testing.T) {
	progress := newRebuildProgress()

	require.NoError(t, progress.start())
	progress.finish(errTestRebuildFailed)

	require.False(t, progress.snapshot().IsRunning)
	require.NoError(t, progress.start())
}

func Test_Book_RebuildFailsFastWhileRunning(t *testing.T) {
	book := newTestBook()

	require.NoError(t, book.progress.start())

	err := book.Rebuild(context.Background())
	require.ErrorIs(t, err, user_errors.ErrAddressesRebuildInProgress)

	err = book.RebuildAsync(context.Background())
	require.ErrorIs(t, err, user_errors.ErrAddressesRebuildInProgress)

	require.True(t, book.RebuildStatus().IsRunning)
}

func Test_RebuildProgress_TotalGrowsOnceServicesAreKnown(t *testing.T) {
	progress := newRebuildProgress()
	require.NoError(t, progress.start())

	progress.addSteps(3)
	require.EqualValues(t, 3, progress.snapshot().TotalSteps)

	progress.step()
	progress.step()

	progress.addSteps(2)

	got := progress.snapshot()
	require.True(t, got.IsRunning)
	require.EqualValues(t, 5, got.TotalSteps)
	require.EqualValues(t, 2, got.DoneSteps)
}

func Test_RebuildProgress_DoneNeverExceedsTotal(t *testing.T) {
	progress := newRebuildProgress()
	require.NoError(t, progress.start())

	progress.addSteps(1)
	progress.step()
	progress.step()

	got := progress.snapshot()
	require.EqualValues(t, 1, got.TotalSteps)
	require.EqualValues(t, 1, got.DoneSteps)
}

func Test_RebuildProgress_SuccessCompletesAndClearsError(t *testing.T) {
	progress := newRebuildProgress()

	require.NoError(t, progress.start())
	progress.addSteps(4)
	progress.step()
	progress.finish(errTestRebuildFailed)

	require.Contains(t, progress.snapshot().LastError, errTestRebuildFailed.Error())

	require.NoError(t, progress.start())
	require.Empty(t, progress.snapshot().LastError)
	require.EqualValues(t, 0, progress.snapshot().TotalSteps)

	progress.addSteps(4)
	progress.finish(nil)

	got := progress.snapshot()
	require.False(t, got.IsRunning)
	require.Empty(t, got.LastError)
	require.EqualValues(t, 4, got.DoneSteps)
	require.EqualValues(t, got.TotalSteps, got.DoneSteps)
}

func Test_RebuildProgress_NilReceiverIsNoop(t *testing.T) {
	var progress *rebuildProgress

	progress.addSteps(2)
	progress.step()
}
