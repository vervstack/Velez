import {describe, expect, it, vi} from 'vitest';
import {fireEvent, render, screen} from '@testing-library/react';

import StepsDialogStepper from '@/dialogs/StepsDialog/StepsDialogStepper.tsx';
import {Step} from '@/dialogs/StepsDialog/Step.ts';

interface TestContext {
    name: string;
}

function FirstScreen() {
    return <div/>;
}

const steps: Step<TestContext>[] = [
    {name: 'first', label: 'First', component: FirstScreen},
    {name: 'second', label: 'Second', component: FirstScreen},
    {name: 'third', label: 'Third', component: FirstScreen},
];

describe('StepsDialogStepper', () => {
    it('shows the first step as active and the rest as upcoming', () => {
        render(
            <StepsDialogStepper steps={steps} currentIndex={0} isLoading={false} onSelect={vi.fn()}/>
        );

        expect(screen.getByText('First').closest('button')).toBeDisabled();
        expect(screen.getByText('Second').closest('button')).toBeDisabled();
        expect(screen.getByText('Third').closest('button')).toBeDisabled();
    });

    it('marks steps before the current index as done and clickable', () => {
        const onSelect = vi.fn();

        render(
            <StepsDialogStepper steps={steps} currentIndex={2} isLoading={false} onSelect={onSelect}/>
        );

        const firstButton = screen.getByText('First').closest('button')!;
        const secondButton = screen.getByText('Second').closest('button')!;
        const thirdButton = screen.getByText('Third').closest('button')!;

        expect(firstButton).not.toBeDisabled();
        expect(secondButton).not.toBeDisabled();
        expect(thirdButton).toBeDisabled();

        fireEvent.click(firstButton);
        expect(onSelect).toHaveBeenCalledWith(0);
    });

    it('marks the rail non-interactive while isLoading', () => {
        const {container} = render(
            <StepsDialogStepper steps={steps} currentIndex={2} isLoading={true} onSelect={vi.fn()}/>
        );

        const rail = container.firstElementChild!;
        expect(rail.className).toMatch(/Disabled/);
    });
});
