import {afterEach, beforeEach, describe, expect, it} from 'vitest';
import {fireEvent, render, screen} from '@testing-library/react';

import {OpenWizardDialog} from '@/dialogs/StepsDialog/openWizardDialog.tsx';
import {StepScreenProps} from '@/dialogs/StepsDialog/Step.ts';
import {useDialog} from '@/app/hooks/dialog/Dialog.tsx';
import Dialog from '@/app/hooks/dialog/Dialog.tsx';

interface TestContext {
    name: string;
}

function FirstScreen({name}: StepScreenProps<TestContext>) {
    return (
        <div>
            <span>First screen</span>
            <span data-testid="name-value">{name}</span>
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
    useDialog.setState({children: null, IsClickOffClosesDialog: true});
});

describe('OpenWizardDialog', () => {
    it('opens a dialog whose children become non-null', () => {
        OpenWizardDialog<TestContext>({
            steps: [
                {name: 'first', component: FirstScreen},
                {name: 'second', component: SecondScreen},
            ],
            initialContext: {name: 'initial'},
        });

        expect(useDialog.getState().children).not.toBeNull();
    });

    it('renders the first step content through the Dialog primitive', () => {
        OpenWizardDialog<TestContext>({
            steps: [
                {name: 'first', component: FirstScreen},
                {name: 'second', component: SecondScreen},
            ],
            initialContext: {name: 'initial'},
        });

        render(<Dialog/>);

        expect(screen.getByText('First screen')).toBeInTheDocument();
        expect(screen.queryByText('Second screen')).not.toBeInTheDocument();
    });

    it('advances to the next step through Next, still driven by the opened wizard', () => {
        OpenWizardDialog<TestContext>({
            steps: [
                {name: 'first', component: FirstScreen},
                {name: 'second', component: SecondScreen},
            ],
            initialContext: {name: 'initial'},
        });

        render(<Dialog/>);

        fireEvent.click(screen.getByText('Next'));

        expect(screen.getByText('Second screen')).toBeInTheDocument();
    });
});
