import {afterEach, describe, expect, it, vi} from 'vitest';
import {renderHook, waitFor} from '@testing-library/react';
import {QueryClient, QueryClientProvider} from '@tanstack/react-query';
import {createElement, ReactNode} from 'react';

import {NodeHardwareQuery} from '@/processes/queries/control_plane.ts';
import {FetchNodeHardware} from '@/processes/api/velez.ts';

vi.mock('@/processes/api/velez.ts', () => ({
    FetchNodeHardware: vi.fn(() => Promise.resolve({isRunningInContainer: true})),
}));

vi.mock('@/processes/api/api.ts', () => ({
    GetInitReq: vi.fn(() => ({pathPrefix: '', headers: {}})),
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

describe('NodeHardwareQuery', () => {
    it('calls through to FetchNodeHardware and returns the resolved data', async () => {
        const {result} = renderHook(() => NodeHardwareQuery(), {wrapper: createWrapper()});

        await waitFor(() => expect(result.current.isSuccess).toBe(true));

        expect(FetchNodeHardware).toHaveBeenCalledTimes(1);
        expect(result.current.data).toEqual({isRunningInContainer: true});
    });
});
