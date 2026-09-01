import {describe, expect, it, vi} from 'vitest';
import {render, screen} from '@testing-library/react';
import {QueryClient, QueryClientProvider} from '@tanstack/react-query';
import {createElement, ReactNode} from 'react';

import Sidebar from '@/widgets/sidebar/Sidebar';

function createWrapper() {
    const queryClient = new QueryClient({defaultOptions: {queries: {retry: false}}});

    return function Wrapper({children}: { children: ReactNode }) {
        return createElement(QueryClientProvider, {client: queryClient}, children);
    };
}

describe('Sidebar', () => {
    it('renders navigation items', () => {
        const Wrapper = createWrapper();

        render(
            createElement(Wrapper, null, createElement(Sidebar, {
                activeNav: 'controlplane',
                onNavChange: vi.fn(),
            })),
        );

        expect(screen.getByText('Control Plane')).toBeInTheDocument();
        expect(screen.getByText('Deployments')).toBeInTheDocument();
    });
});
