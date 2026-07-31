import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
import {fireEvent, render, screen} from '@testing-library/react';

import StepsDialog from '@/dialogs/StepsDialog/StepsDialog.tsx';
import {StepScreenProps} from '@/dialogs/StepsDialog/Step.ts';
import {useDialog} from '@/app/hooks/dialog/Dialog.tsx';
import Button from '@/components/base/Button.tsx';

interface TestContext {
    name: string;
}

function FirstScreen({name, updateContext}: StepScreenProps<TestContext>) {
    return (
        <div>
            <span>First screen</span>
            <span data-testid="name-value">{name}</span>
            <Button onClick={() => updateContext({name: 'updated-by-step-one'})}>
                Set name
            </Button>
        </div>
    );
}

function SecondScreen({name}: StepScreenProps<TestContext>) {
    return (
        <div>
            <span>Second screen</span>
            <span data-testid="name-value">{name}</span>
        </div>
    );
}

function TestFinalStep({context, onClose}: {context: TestContext; onClose(): void}) {
    return (
        <div>
            <span>Final screen</span>
            <span data-testid="final-name-value">{context.name}</span>
            <button onClick={onClose}>Close final</button>
        </div>
    );
}

// Captured once, before any test spies on it — zustand's setState merges by
// creating a new state object each call, so a spy installed via
// `vi.spyOn(useDialog.getState(), 'CloseDialog')` gets copied by reference
// into every subsequent merged state. `vi.restoreAllMocks()` only restores
// the property on the specific (now-stale) object it was originally taken
// from, not the store's current object, so the mocked function otherwise
// leaks into later tests. Forcing CloseDialog back to this pristine
// reference before each test keeps every test's spy independent.
const pristineCloseDialog = useDialog.getState().CloseDialog;

beforeEach(() => {
    useDialog.setState({children: null, IsClickOffClosesDialog: true, CloseDialog: pristineCloseDialog});
});

afterEach(() => {
    vi.restoreAllMocks();
});

describe('StepsDialog', () => {
    it('renders the first step initially', () => {
        render(
            <StepsDialog<TestContext>
                steps={[
                    {name: 'first', component: FirstScreen},
                    {name: 'second', component: SecondScreen},
                ]}
                initialContext={{name: 'initial'}}
            />
        );

        expect(screen.getByText('First screen')).toBeInTheDocument();
        expect(screen.queryByText('Second screen')).not.toBeInTheDocument();
    });

    it('advances to the next step and returns with Back', () => {
        render(
            <StepsDialog<TestContext>
                steps={[
                    {name: 'first', component: FirstScreen},
                    {name: 'second', component: SecondScreen},
                ]}
                initialContext={{name: 'initial'}}
            />
        );

        fireEvent.click(screen.getByText('Next'));
        expect(screen.getByText('Second screen')).toBeInTheDocument();

        fireEvent.click(screen.getByText('Back'));
        expect(screen.getByText('First screen')).toBeInTheDocument();
    });

    it('keeps a context value set in step one visible in step two', () => {
        render(
            <StepsDialog<TestContext>
                steps={[
                    {name: 'first', component: FirstScreen},
                    {name: 'second', component: SecondScreen},
                ]}
                initialContext={{name: 'initial'}}
            />
        );

        fireEvent.click(screen.getByText('Set name'));
        fireEvent.click(screen.getByText('Next'));

        expect(screen.getByTestId('name-value')).toHaveTextContent('updated-by-step-one');
    });

    it('disables Next when the current step cannot proceed', () => {
        render(
            <StepsDialog<TestContext>
                steps={[
                    {name: 'first', component: FirstScreen, canProceed: () => false},
                    {name: 'second', component: SecondScreen},
                ]}
                initialContext={{name: 'initial'}}
            />
        );

        expect(screen.getByText('Next')).toBeDisabled();
    });

    it('jumps back to a completed step when its stepper row is clicked', () => {
        render(
            <StepsDialog<TestContext>
                steps={[
                    {name: 'first', label: 'First', component: FirstScreen},
                    {name: 'second', label: 'Second', component: SecondScreen},
                ]}
                initialContext={{name: 'initial'}}
            />
        );

        fireEvent.click(screen.getByText('Next'));
        expect(screen.getByText('Second screen')).toBeInTheDocument();

        fireEvent.click(screen.getByText('First').closest('button')!);
        expect(screen.getByText('First screen')).toBeInTheDocument();
    });

    it('does not navigate when clicking the active step in the stepper', () => {
        render(
            <StepsDialog<TestContext>
                steps={[
                    {name: 'first', label: 'First', component: FirstScreen},
                    {name: 'second', label: 'Second', component: SecondScreen},
                ]}
                initialContext={{name: 'initial'}}
            />
        );

        fireEvent.click(screen.getByText('First').closest('button')!);
        expect(screen.getByText('First screen')).toBeInTheDocument();
    });

    it('calls onFinish with the final context and closes the dialog on Finish', () => {
        const onFinish = vi.fn();
        const closeSpy = vi.spyOn(useDialog.getState(), 'CloseDialog');

        render(
            <StepsDialog<TestContext>
                steps={[
                    {name: 'first', component: FirstScreen},
                    {name: 'second', component: SecondScreen},
                ]}
                initialContext={{name: 'initial'}}
                onFinish={onFinish}
            />
        );

        fireEvent.click(screen.getByText('Set name'));
        fireEvent.click(screen.getByText('Next'));
        fireEvent.click(screen.getByText('Finish'));

        expect(onFinish).toHaveBeenCalledWith({name: 'updated-by-step-one'});
        expect(closeSpy).toHaveBeenCalledTimes(1);
    });

    it('renders finalStep instead of closing when Finish is clicked and finalStep is provided', () => {
        const onFinish = vi.fn();
        const closeSpy = vi.spyOn(useDialog.getState(), 'CloseDialog');

        render(
            <StepsDialog<TestContext>
                steps={[
                    {name: 'first', label: 'First', component: FirstScreen},
                    {name: 'second', label: 'Second', component: SecondScreen},
                ]}
                initialContext={{name: 'initial'}}
                onFinish={onFinish}
                finalStep={TestFinalStep}
            />
        );

        fireEvent.click(screen.getByText('Set name'));
        fireEvent.click(screen.getByText('Next'));
        fireEvent.click(screen.getByText('Finish'));

        expect(onFinish).not.toHaveBeenCalled();
        expect(closeSpy).not.toHaveBeenCalled();

        expect(screen.getByText('Final screen')).toBeInTheDocument();
        expect(screen.getByTestId('final-name-value')).toHaveTextContent('updated-by-step-one');

        expect(screen.queryByText('Second screen')).not.toBeInTheDocument();
        expect(screen.queryByText('Back')).not.toBeInTheDocument();
        expect(screen.queryByText('Finish')).not.toBeInTheDocument();
    });

    it('keeps the header visible while showing finalStep', () => {
        render(
            <StepsDialog<TestContext>
                steps={[
                    {name: 'first', label: 'First', component: FirstScreen},
                ]}
                initialContext={{name: 'initial'}}
                header={{icon: '⛁', eyebrow: 'Plugins', eyebrowDetail: 'test'}}
                finalStep={TestFinalStep}
            />
        );

        fireEvent.click(screen.getByText('Finish'));

        expect(screen.getByText('Plugins')).toBeInTheDocument();
        expect(screen.getByText('test')).toBeInTheDocument();
        expect(screen.getByText('Final screen')).toBeInTheDocument();
    });

    it('calls onClose (CloseDialog) when the finalStep close affordance is used', () => {
        const closeSpy = vi.spyOn(useDialog.getState(), 'CloseDialog');

        render(
            <StepsDialog<TestContext>
                steps={[
                    {name: 'first', label: 'First', component: FirstScreen},
                ]}
                initialContext={{name: 'initial'}}
                finalStep={TestFinalStep}
            />
        );

        fireEvent.click(screen.getByText('Finish'));
        fireEvent.click(screen.getByText('Close final'));

        expect(closeSpy).toHaveBeenCalledTimes(1);
    });
});
