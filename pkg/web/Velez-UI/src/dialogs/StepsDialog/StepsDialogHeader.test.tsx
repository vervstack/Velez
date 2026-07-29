import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {render, screen} from '@testing-library/react';

import StepsDialogHeader from '@/dialogs/StepsDialog/StepsDialogHeader.tsx';

beforeEach(() => {
    vi.spyOn(console, 'error').mockImplementation(() => {});
});

afterEach(() => {
    vi.restoreAllMocks();
});

describe('StepsDialogHeader', () => {
    it('renders real header content without skeletons or console.error when not loading', () => {
        render(
            <StepsDialogHeader
                header={{icon: '⛁', eyebrow: 'Plugins', eyebrowDetail: 'statefull-pg'}}
                isLoading={false}
                stepLabel="Overview"
                stepIndex={0}
                totalSteps={3}
                onClose={() => {}}
            />
        );

        expect(screen.getByText('⛁')).toBeInTheDocument();
        expect(screen.getByText('Plugins')).toBeInTheDocument();
        expect(screen.getByText('statefull-pg')).toBeInTheDocument();
        expect(console.error).not.toHaveBeenCalled();
    });

    it('renders skeleton placeholders without console.error while loading', () => {
        const {container} = render(
            <StepsDialogHeader
                header={{}}
                isLoading={true}
                stepLabel="Overview"
                stepIndex={0}
                totalSteps={3}
                onClose={() => {}}
            />
        );

        expect(container.querySelectorAll('[class*="Skeleton"]').length).toBeGreaterThan(0);
        expect(console.error).not.toHaveBeenCalled();
    });

    it('renders skeleton placeholders as a fallback and logs console.error when not loading but fields are missing', () => {
        const {container} = render(
            <StepsDialogHeader
                header={{}}
                isLoading={false}
                stepLabel="Overview"
                stepIndex={0}
                totalSteps={3}
                onClose={() => {}}
            />
        );

        expect(container.querySelectorAll('[class*="Skeleton"]').length).toBeGreaterThan(0);
        expect(console.error).toHaveBeenCalled();
    });
});
