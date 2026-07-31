import {afterEach, describe, expect, it, vi} from 'vitest';
import {render, screen, waitFor} from '@testing-library/react';
import {QueryClient, QueryClientProvider} from '@tanstack/react-query';
import {createElement, ReactNode} from 'react';

import Sidebar from '@/widgets/sidebar/Sidebar';
import {NodeStatus, VervPluginState, VervPluginType} from '@/app/api/velez';
import {VervPlugin} from '@/model/services/VervPlugins.tsx';

const statefullPlugin = new VervPlugin(VervPluginType.statefull_pg, 'pg');
statefullPlugin.state = VervPluginState.running;

vi.mock('@/processes/api/control_plane', () => ({
    controlPlaneService: {
        listPlugins: vi.fn(() => Promise.resolve([statefullPlugin])),
        listNodes: vi.fn(() => Promise.resolve({
            nodes: [{id: 'node-1', name: 'node-1', addr: '10.0.0.1', status: NodeStatus.NodeStatus_Online}],
        })),
    },
}));

afterEach(() => {
    vi.restoreAllMocks();
});

function createWrapper() {
    const queryClient = new QueryClient({defaultOptions: {queries: {retry: false}}});

    return function Wrapper({children}: { children: ReactNode }) {
        return createElement(QueryClientProvider, {client: queryClient}, children);
    };
}

describe('Sidebar', () => {
    it('does not throw when the statefull-mode plugin query resolves after initial render', async () => {
        const Wrapper = createWrapper();

        render(
            createElement(Wrapper, null, createElement(Sidebar, {
                collapsed: false,
                onNodeSelect: vi.fn(),
                activeNav: 'controlplane',
                onNavChange: vi.fn(),
            })),
        );

        await waitFor(() => expect(screen.getByText('Nodes')).toBeInTheDocument());
        expect(screen.getByText('node-1')).toBeInTheDocument();
    });
});
