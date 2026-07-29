import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {render, screen} from '@testing-library/react';

import StepsDialog from '@/dialogs/StepsDialog/StepsDialog.tsx';
import {StepScreenProps} from '@/dialogs/StepsDialog/Step.ts';
import {useDialog} from '@/app/hooks/dialog/Dialog.tsx';

interface TestContext {
    name: string;
}

function FirstScreen({name}: StepScreenProps<TestContext>) {
    return <span>First screen content {name}</span>;
}

beforeEach(() => {
    useDialog.setState({children: null, IsClickOffClosesDialog: true});
});

afterEach(() => {
    vi.restoreAllMocks();
});

describe('StepsDialog loading state', () => {
    it('renders the screen skeleton loader instead of the step content when isLoading is true', () => {
        render(
            <StepsDialog<TestContext>
                steps={[{name: 'first', component: FirstScreen}]}
                initialContext={{name: 'initial'}}
                isLoading={true}
            />
        );

        expect(screen.queryByText(/First screen content/)).not.toBeInTheDocument();
        expect(screen.getByTestId('screen-skeleton-loader')).toBeInTheDocument();
    });

    it('renders the loading hint when loadingHint is passed', () => {
        render(
            <StepsDialog<TestContext>
                steps={[{name: 'first', component: FirstScreen}]}
                initialContext={{name: 'initial'}}
                isLoading={true}
                loadingHint="Loading this node's hardware to configure cluster mode…"
            />
        );

        expect(screen.getByText("Loading this node's hardware to configure cluster mode…")).toBeInTheDocument();
    });

    it('does not render a loading hint when loadingHint is not passed', () => {
        render(
            <StepsDialog<TestContext>
                steps={[{name: 'first', component: FirstScreen}]}
                initialContext={{name: 'initial'}}
                isLoading={true}
            />
        );

        expect(screen.getByTestId('screen-skeleton-loader').querySelector('p')).not.toBeInTheDocument();
    });

    it('renders the step content instead of the skeleton loader when isLoading is false', () => {
        render(
            <StepsDialog<TestContext>
                steps={[{name: 'first', component: FirstScreen}]}
                initialContext={{name: 'initial'}}
                isLoading={false}
            />
        );

        expect(screen.getByText(/First screen content/)).toBeInTheDocument();
        expect(screen.queryByTestId('screen-skeleton-loader')).not.toBeInTheDocument();
    });
});
