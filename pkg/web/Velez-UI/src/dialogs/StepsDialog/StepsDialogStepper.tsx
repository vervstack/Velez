import cn from 'classnames';

import {Step} from '@/dialogs/StepsDialog/Step.ts';
import cls from '@/dialogs/StepsDialog/StepsDialogStepper.module.css';

interface StepsDialogStepperProps<TContext> {
    steps: Step<TContext>[];
    currentIndex: number;
    isLoading: boolean;

    onSelect(index: number): void;
}

function padIndex(index: number): string {
    return String(index + 1).padStart(2, '0');
}

export default function StepsDialogStepper<TContext>(
    {steps, currentIndex, isLoading, onSelect}: StepsDialogStepperProps<TContext>) {
    return (
        <div className={cn(cls.StepsDialogStepperContainer, isLoading && cls.Disabled)}>
            <div className={cls.RailLabel}>Pipeline</div>

            {steps.map((step, index) => {
                const isDone = index < currentIndex;
                const isActive = index === currentIndex;
                const isUpcoming = index > currentIndex;
                const isLast = index === steps.length - 1;

                return (
                    <button
                        key={step.name}
                        type="button"
                        className={cls.StepRow}
                        disabled={!isDone}
                        onClick={() => onSelect(index)}
                    >
                        <div
                            className={cn(
                                cls.StepDot,
                                isDone && cls.StepDotDone,
                                isActive && cls.StepDotActive,
                                isUpcoming && cls.StepDotUpcoming,
                                isLast && cls.StepDotLast
                            )}
                        >
                            {padIndex(index)}
                            {isDone && <span className={cls.StepDotCheck}>✓</span>}
                        </div>

                        <span
                            className={cn(
                                cls.StepLabel,
                                isDone && cls.StepLabelDone,
                                isActive && cls.StepLabelActive,
                                isUpcoming && cls.StepLabelUpcoming
                            )}
                        >
                            {step.label ?? step.name}
                        </span>
                    </button>
                );
            })}
        </div>
    );
}
