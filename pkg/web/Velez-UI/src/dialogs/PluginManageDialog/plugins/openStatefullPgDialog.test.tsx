import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {render, screen} from '@testing-library/react';

import {VervPluginState, VervPluginType} from '@/app/api/velez';
import {useDialog} from '@/app/hooks/dialog/Dialog.tsx';
import Dialog from '@/app/hooks/dialog/Dialog.tsx';
import {queryClient} from '@/app/queryClient.ts';
import {VervPlugin} from '@/model/services/VervPlugins.tsx';
import {openStatefullPgDialog} from '@/dialogs/PluginManageDialog/plugins/openStatefullPgDialog.tsx';

vi.mock('@/processes/api/velez.ts', () => ({
    FetchNodeHardware: vi.fn(() => Promise.resolve({isRunningInContainer: true})),
}));

vi.mock('@/processes/api/control_plane.ts', () => ({
    controlPlaneService: {
        enableStatefullPgCluster: vi.fn(() => Promise.resolve()),
    },
}));

beforeEach(() => {
    useDialog.setState({children: null, IsClickOffClosesDialog: true});
    localStorage.setItem('velez_api_key', 'test-key');
    queryClient.clear();
});

afterEach(() => {
    vi.restoreAllMocks();
    localStorage.clear();
    queryClient.clear();
});

describe('openStatefullPgDialog', () => {
    it('opens the dead prompt when the plugin is dead', async () => {
        const plugin = new VervPlugin(VervPluginType.statefull_pg, 'statefull-pg-service');
        plugin.state = VervPluginState.dead;

        await openStatefullPgDialog(plugin);

        expect(useDialog.getState().children).not.toBeNull();
    });

    it('opens the wizard dialog when the plugin is not dead/running', async () => {
        const plugin = new VervPlugin(VervPluginType.statefull_pg, 'statefull-pg-service');
        plugin.state = VervPluginState.disabled;

        await openStatefullPgDialog(plugin);

        expect(useDialog.getState().children).not.toBeNull();
    });

    it('opens the dialog synchronously in a loading state before the hardware fetch resolves', async () => {
        const plugin = new VervPlugin(VervPluginType.statefull_pg, 'statefull-pg-service');
        plugin.state = VervPluginState.disabled;

        const resultPromise = openStatefullPgDialog(plugin);

        expect(useDialog.getState().children).not.toBeNull();

        await resultPromise;
    });

    it('passes the loading hint to the dialog tree while the hardware fetch is pending', async () => {
        const plugin = new VervPlugin(VervPluginType.statefull_pg, 'statefull-pg-service');
        plugin.state = VervPluginState.disabled;

        const resultPromise = openStatefullPgDialog(plugin);

        render(<Dialog/>);

        expect(screen.getByText("Loading this node's hardware to configure cluster mode…")).toBeInTheDocument();

        await resultPromise;
    });
});
