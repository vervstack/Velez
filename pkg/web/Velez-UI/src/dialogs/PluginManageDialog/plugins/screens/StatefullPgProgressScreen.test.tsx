import {createElement, ReactNode} from 'react';
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {fireEvent, render, screen, waitFor} from '@testing-library/react';
import {QueryClient, QueryClientProvider} from '@tanstack/react-query';

import {TaskStatus, TaskStatusStatus} from '@/app/api/velez';
import StatefullPgProgressScreen from '@/dialogs/PluginManageDialog/plugins/screens/StatefullPgProgressScreen.tsx';
import {StatefullPgContext} from '@/dialogs/PluginManageDialog/plugins/StatefullPgContext.ts';
import {controlPlaneService} from '@/processes/api/control_plane.ts';
import {WatchTaskStream} from '@/processes/api/tasks.ts';

vi.mock('@/processes/api/control_plane.ts', () => ({
    controlPlaneService: {
        enableStatefullPgCluster: vi.fn(),
    },
}));

vi.mock('@/processes/api/tasks.ts', () => ({
    WatchTaskStream: vi.fn(),
}));

vi.mock('@/processes/api/velez.ts', () => ({
    FetchNodeHardware: vi.fn(() => Promise.resolve({nodeRegion: 'eu-west-1'})),
}));

const context: StatefullPgContext = {exposePort: true, portNumber: '5432', isRunningInContainer: true};

const RUNNING_JOBS = [
    {name: 'generate_credentials', status: TaskStatusStatus.DONE},
    {name: 'create_container', status: TaskStatusStatus.DONE},
    {name: 'start_container', status: TaskStatusStatus.RUNNING},
    {name: 'wait_for_postgres_ready', status: TaskStatusStatus.PENDING},
];

const FAILED_JOBS = [
    {name: 'generate_credentials', status: TaskStatusStatus.DONE},
    {name: 'create_container', status: TaskStatusStatus.FAILED},
    {name: 'start_container', status: TaskStatusStatus.PENDING},
];

function renderWithClient(ui: ReactNode) {
    const queryClient = new QueryClient({defaultOptions: {queries: {retry: false}}});

    return render(createElement(QueryClientProvider, {client: queryClient}, ui));
}

beforeEach(() => {
    vi.mocked(controlPlaneService.enableStatefullPgCluster)
        .mockResolvedValue({entityId: 'entity-1', action: 'enable_statefull_pg'});
});

afterEach(() => {
    vi.restoreAllMocks();
});

describe('StatefullPgProgressScreen', () => {
    it('goes from pending to success once the watched task completes, showing the task id and step chips', async () => {
        vi.mocked(WatchTaskStream).mockImplementation(async (_req, onStatus) => {
            onStatus({
                status: TaskStatusStatus.RUNNING,
                taskId: 'task-42',
                jobs: RUNNING_JOBS,
            } as TaskStatus);
            onStatus({
                status: TaskStatusStatus.DONE,
                taskId: 'task-42',
                jobs: RUNNING_JOBS.map((job) => ({...job, status: TaskStatusStatus.DONE})),
            } as TaskStatus);
        });

        const {container} = renderWithClient(
            <StatefullPgProgressScreen context={context} onClose={vi.fn()}/>
        );

        await waitFor(() => {
            expect(screen.getByText('Enabling cluster mode · Task #task-42')).toBeInTheDocument();
        });

        expect(screen.getByText('Generating credentials')).toBeInTheDocument();
        expect(screen.getByText('Starting container')).toBeInTheDocument();
        expect(screen.getByText('Waiting for Postgres')).toBeInTheDocument();

        await waitFor(() => {
            expect(screen.getByText('Done')).toBeInTheDocument();
        });

        expect(controlPlaneService.enableStatefullPgCluster).toHaveBeenCalledWith({
            isExposePort: true,
            exposeToPort: '5432',
        });
        expect(WatchTaskStream).toHaveBeenCalledWith(
            {entityId: 'entity-1', action: 'enable_statefull_pg'},
            expect.any(Function)
        );

        expect(container.querySelectorAll('[class*="done"]').length).toBeGreaterThan(0);
    });

    it('renders done/running/pending step chips as visually distinct while running', async () => {
        vi.mocked(WatchTaskStream).mockImplementation(async (_req, onStatus) => {
            onStatus({
                status: TaskStatusStatus.RUNNING,
                taskId: 'task-7',
                jobs: RUNNING_JOBS,
            } as TaskStatus);
        });

        renderWithClient(
            <StatefullPgProgressScreen context={context} onClose={vi.fn()}/>
        );

        await waitFor(() => {
            expect(screen.getByText('Starting container')).toBeInTheDocument();
        });

        const doneChip = screen.getByText('Creating container').closest('[class*="ProgressStepChipContainer"]');
        const runningChip = screen.getByText('Starting container').closest('[class*="ProgressStepChipContainer"]');
        const pendingChip = screen.getByText('Waiting for Postgres').closest('[class*="ProgressStepChipContainer"]');

        expect(doneChip?.className).toMatch(/done/);
        expect(runningChip?.className).toMatch(/running/);
        expect(pendingChip?.className).not.toMatch(/done|running|failed/);

        expect(runningChip?.querySelector('[class*="Shimmer"]')).not.toBeNull();
        expect(doneChip?.querySelector('[class*="Shimmer"]')).toBeNull();
    });

    it('goes from pending to failed and shows the error message and a failed step chip', async () => {
        vi.mocked(WatchTaskStream).mockImplementation(async (_req, onStatus) => {
            onStatus({
                status: TaskStatusStatus.FAILED,
                error: 'disk full',
                taskId: 'task-9',
                jobs: FAILED_JOBS,
            } as TaskStatus);
        });

        renderWithClient(<StatefullPgProgressScreen context={context} onClose={vi.fn()}/>);

        await waitFor(() => {
            expect(screen.getByText('disk full')).toBeInTheDocument();
        });

        const failedChip = screen.getByText('Creating container').closest('[class*="ProgressStepChipContainer"]');
        expect(failedChip?.className).toMatch(/failed/);
    });

    it('calls onClose when Done is clicked after success', async () => {
        vi.mocked(WatchTaskStream).mockImplementation(async (_req, onStatus) => {
            onStatus({status: TaskStatusStatus.DONE, taskId: 'task-1', jobs: []} as TaskStatus);
        });
        const onClose = vi.fn();

        renderWithClient(<StatefullPgProgressScreen context={context} onClose={onClose}/>);

        const doneButton = await screen.findByText('Done');
        fireEvent.click(doneButton);

        expect(onClose).toHaveBeenCalledTimes(1);
    });

    it('calls onClose when Close is clicked after failure', async () => {
        vi.mocked(WatchTaskStream).mockImplementation(async (_req, onStatus) => {
            onStatus({status: TaskStatusStatus.FAILED, error: 'boom', taskId: 'task-2', jobs: []} as TaskStatus);
        });
        const onClose = vi.fn();

        renderWithClient(<StatefullPgProgressScreen context={context} onClose={onClose}/>);

        const closeButton = await screen.findByText('Close');
        fireEvent.click(closeButton);

        expect(onClose).toHaveBeenCalledTimes(1);
    });
});
