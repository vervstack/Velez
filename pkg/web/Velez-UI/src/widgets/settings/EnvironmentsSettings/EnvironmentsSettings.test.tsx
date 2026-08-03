import {afterEach, describe, expect, it, vi} from 'vitest';
import {fireEvent, render, screen, waitFor} from '@testing-library/react';
import {QueryClient, QueryClientProvider} from '@tanstack/react-query';
import {createElement, ReactNode} from 'react';

import EnvironmentsSettings from '@/widgets/settings/EnvironmentsSettings/EnvironmentsSettings.tsx';
import Dialog from '@/app/hooks/dialog/Dialog.tsx';
import {VervPluginState, VervPluginType} from '@/app/api/velez';
import {VervPlugin} from '@/model/services/VervPlugins.tsx';

const statefullPlugin = new VervPlugin(VervPluginType.statefull_pg, 'pg');
statefullPlugin.state = VervPluginState.running;

const environments = [
    {id: '1', name: 'PROD', suffix: ''},
];

const listEnvironments = vi.fn(() => Promise.resolve({environments}));
const createEnvironment = vi.fn((_name: string, _suffix?: string) =>
    Promise.resolve({environment: {id: '2', name: 'STAGING'}}));
const updateEnvironment = vi.fn((_id: string, _name: string, _suffix?: string) =>
    Promise.resolve({environment: {id: '1', name: 'PROD'}}));
const deleteEnvironment = vi.fn((_id: string) => Promise.resolve());

vi.mock('@/processes/api/control_plane', () => ({
    controlPlaneService: {
        listPlugins: vi.fn(() => Promise.resolve([statefullPlugin])),
        listEnvironments: () => listEnvironments(),
        createEnvironment: (name: string, suffix?: string) => createEnvironment(name, suffix),
        updateEnvironment: (id: string, name: string, suffix?: string) => updateEnvironment(id, name, suffix),
        deleteEnvironment: (id: string) => deleteEnvironment(id),
    },
}));

afterEach(() => {
    vi.clearAllMocks();
});

function renderWithProviders() {
    const queryClient = new QueryClient({defaultOptions: {queries: {retry: false}}});

    function Wrapper({children}: { children: ReactNode }) {
        return createElement(QueryClientProvider, {client: queryClient}, children);
    }

    return render(
        createElement(Wrapper, null,
            createElement('div', null,
                createElement(EnvironmentsSettings),
                createElement(Dialog),
            ),
        ),
    );
}

describe('EnvironmentsSettings', () => {
    it('lists environments fetched from ListEnvironmentsQuery', async () => {
        renderWithProviders();

        await waitFor(() => expect(screen.getByText('PROD')).toBeInTheDocument());
    });

    it('creates a new environment via the Add environment dialog', async () => {
        renderWithProviders();

        await waitFor(() => expect(screen.getByText('PROD')).toBeInTheDocument());

        fireEvent.click(screen.getByText('+ Add environment'));

        const nameInput = screen.getAllByRole('textbox')[0];
        fireEvent.change(nameInput, {target: {value: 'STAGING'}});

        fireEvent.click(screen.getByText('Create'));

        await waitFor(() => expect(createEnvironment).toHaveBeenCalledWith('STAGING', undefined));
    });

    it('updates an environment via the Manage dialog', async () => {
        renderWithProviders();

        await waitFor(() => expect(screen.getByText('PROD')).toBeInTheDocument());

        fireEvent.click(screen.getByText('Manage'));

        const nameInput = screen.getAllByRole('textbox')[0];
        fireEvent.change(nameInput, {target: {value: 'PRODUCTION'}});

        fireEvent.click(screen.getByText('Save'));

        await waitFor(() => expect(updateEnvironment).toHaveBeenCalledWith('1', 'PRODUCTION', undefined));
    });

    it('deletes an environment via the Manage dialog delete confirmation', async () => {
        renderWithProviders();

        await waitFor(() => expect(screen.getByText('PROD')).toBeInTheDocument());

        fireEvent.click(screen.getByText('Manage'));
        fireEvent.click(screen.getByText('Delete'));
        fireEvent.click(screen.getByText('Remove'));

        await waitFor(() => expect(deleteEnvironment).toHaveBeenCalledWith('1'));
    });
});
