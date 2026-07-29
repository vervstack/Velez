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
    it('keeps rendering the step content while isLoading is true', () => {
        render(
            <StepsDialog<TestContext>
                steps={[{name: 'first', component: FirstScreen}]}
                initialContext={{name: 'initial'}}
                isLoading={true}
            />
        );

        expect(screen.getByText(/First screen content/)).toBeInTheDocument();
    });

    it('disables Next while isLoading is true', () => {
        render(
            <StepsDialog<TestContext>
                steps={[{name: 'first', component: FirstScreen}]}
                initialContext={{name: 'initial'}}
                isLoading={true}
            />
        );

        expect(screen.getByText('Finish')).toBeDisabled();
    });

    it('enables Next once isLoading is false', () => {
        render(
            <StepsDialog<TestContext>
                steps={[{name: 'first', component: FirstScreen}]}
                initialContext={{name: 'initial'}}
                isLoading={false}
            />
        );

        expect(screen.getByText('Finish')).not.toBeDisabled();
    });

    it('shows the loading hint next to the Next button when isLoading is true', () => {
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

        expect(screen.queryByText(/Loading/)).not.toBeInTheDocument();
    });

    it('does not render the loading hint once isLoading is false', () => {
        render(
            <StepsDialog<TestContext>
                steps={[{name: 'first', component: FirstScreen}]}
                initialContext={{name: 'initial'}}
                isLoading={false}
                loadingHint="Loading this node's hardware to configure cluster mode…"
            />
        );

        expect(screen.queryByText(/Loading this node's hardware/)).not.toBeInTheDocument();
    });
});
