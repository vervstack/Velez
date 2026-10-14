import {beforeEach, describe, expect, it, vi} from 'vitest';
import {render, screen} from '@testing-library/react';

import SidebarVersion from '@/widgets/sidebar/SidebarVersion';
import {useVersionQuery} from '@/processes/queries/version.ts';

vi.mock('@/processes/queries/version.ts', () => ({useVersionQuery: vi.fn()}));

type VersionQuery = ReturnType<typeof useVersionQuery>;

function renderFooter(query: Partial<VersionQuery>) {
    vi.mocked(useVersionQuery).mockReturnValue(query as Partial<VersionQuery> as VersionQuery);

    render(<SidebarVersion/>);
}

describe('SidebarVersion', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('shows the version as returned when loaded', () => {
        renderFooter({isLoading: false, isError: false, data: {version: 'v1.2.3'}});

        expect(screen.getByText('Velez · v1.2.3')).toBeInTheDocument();
    });

    it('shows unknown when the request fails', () => {
        renderFooter({isLoading: false, isError: true, data: undefined});

        expect(screen.getByText('Velez · unknown')).toBeInTheDocument();
    });

    it('hides the label while loading', () => {
        renderFooter({isLoading: true, isError: false, data: undefined});

        expect(screen.queryByText(/Velez ·/)).not.toBeInTheDocument();
    });
});
