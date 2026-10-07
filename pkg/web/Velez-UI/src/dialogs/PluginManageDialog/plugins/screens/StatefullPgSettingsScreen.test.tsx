import {describe, expect, it, vi} from 'vitest';
import {fireEvent, render, screen} from '@testing-library/react';

import StatefullPgSettingsScreen from '@/dialogs/PluginManageDialog/plugins/screens/StatefullPgSettingsScreen.tsx';

describe('StatefullPgSettingsScreen', () => {
    it('forces the expose-port checkbox on and disables it when running as a binary', () => {
        const updateContext = vi.fn();

        render(
            <StatefullPgSettingsScreen
                exposePort={true}
                portNumber={'5432'}
                isRunningInContainer={false}
                updateContext={updateContext}
            />
        );

        expect(screen.getByText(/must stay exposed/i)).toBeInTheDocument();
        expect(screen.getByRole('checkbox', {name: 'Expose port'})).toBeDisabled();
    });

    it('lets the user toggle expose-port freely when running in a container', () => {
        const updateContext = vi.fn();

        render(
            <StatefullPgSettingsScreen
                exposePort={false}
                portNumber={'5432'}
                isRunningInContainer={true}
                updateContext={updateContext}
            />
        );

        expect(screen.queryByText(/must stay exposed/i)).not.toBeInTheDocument();

        const checkbox = screen.getByRole('checkbox', {name: 'Expose port'});
        expect(checkbox).not.toBeDisabled();

        fireEvent.click(checkbox);

        expect(updateContext).toHaveBeenCalledWith({exposePort: true});
    });
});
