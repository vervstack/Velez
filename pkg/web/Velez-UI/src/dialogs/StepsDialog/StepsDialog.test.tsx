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

beforeEach(() => {
    useDialog.setState({children: null, IsClickOffClosesDialog: true});
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
});
