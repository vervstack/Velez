import {describe, expect, it, vi} from 'vitest';
import {fireEvent, render, screen} from '@testing-library/react';

import DeploymentFilters from '@/widgets/deployments/DeploymentFilters';

function renderFilters(overrides: Partial<Parameters<typeof DeploymentFilters>[0]> = {}) {
    const props = {
        search: '',
        onSearchChange: vi.fn(),
        statusFilters: new Set<string>(),
        onToggleStatus: vi.fn(),
        envFilters: new Set<string>(),
        onToggleEnv: vi.fn(),
        onClearAll: vi.fn(),
        viewMode: 'kanban' as const,
        onViewModeChange: vi.fn(),
        totalCount: 0,
        ...overrides,
    };

    return {
        props,
        ...render(<DeploymentFilters {...props} />),
    };
}

describe('DeploymentFilters', () => {
    it('renders the clear button when a filter is active', () => {
        renderFilters({search: 'foo'});

        expect(screen.getByText('✕ clear')).toBeInTheDocument();
    });

    it('does not render the clear button when no filters are active', () => {
        renderFilters();

        expect(screen.queryByText('✕ clear')).not.toBeInTheDocument();
    });

    it('keeps the clear button mounted during its exit animation, then unmounts after animationend', () => {
        const {rerender, props} = renderFilters({search: 'foo'});

        rerender(<DeploymentFilters {...props} search="" />);

        const button = screen.getByText('✕ clear');
        expect(button).toBeInTheDocument();

        // jsdom doesn't implement the CSS Animations API, so React's vendor-prefix
        // feature detection (getVendorPrefixedEventName) resolves 'animationend' to
        // 'webkitAnimationEnd' in this environment. fireEvent.animationEnd() dispatches
        // the unprefixed native event and never reaches onAnimationEnd here, so dispatch
        // the prefixed event name React actually listens for.
        fireEvent(button, new Event('webkitAnimationEnd', {bubbles: true, cancelable: true}));

        expect(screen.queryByText('✕ clear')).not.toBeInTheDocument();
    });

    it('calls onClearAll when the clear button is clicked', () => {
        const onClearAll = vi.fn();
        renderFilters({search: 'foo', onClearAll});

        fireEvent.click(screen.getByText('✕ clear'));

        expect(onClearAll).toHaveBeenCalledTimes(1);
    });
});
