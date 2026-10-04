package local_storage

import (
	"context"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/local_state"
	"go.vervstack.ru/Velez/internal/domain"
)

// localSettings is the single-node/dev storage.SettingsStorage: there is no
// velez.settings table, so the toggles live in the local state file.
type localSettings struct {
	state node_clients.StateManager
}

func newSettingsStorage(state node_clients.StateManager) *localSettings {
	return &localSettings{state: state}
}

func (l *localSettings) GetSettings(_ context.Context) (domain.Settings, error) {
	current := l.state.Get().Settings

	settings := domain.Settings{
		IsSysboxEnabled:          current.IsSysboxEnabled,
		IsSysboxWhitelistIgnored: current.IsSysboxWhitelistIgnored,
	}

	return settings, nil
}

func (l *localSettings) UpdateSettings(_ context.Context, settings domain.Settings) (domain.Settings, error) {
	state := l.state.GetForUpdate()

	state.Settings = local_state.Settings{
		IsSysboxEnabled:          settings.IsSysboxEnabled,
		IsSysboxWhitelistIgnored: settings.IsSysboxWhitelistIgnored,
	}

	l.state.SetAndRelease(state)

	return settings, nil
}
