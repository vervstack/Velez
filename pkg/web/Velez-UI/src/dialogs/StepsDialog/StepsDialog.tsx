import {useState} from 'react';

import {useDialog} from '@/app/hooks/dialog/Dialog.tsx';
import StepsDialogHeader from '@/dialogs/StepsDialog/StepsDialogHeader.tsx';
import StepsDialogFooter from '@/dialogs/StepsDialog/StepsDialogFooter.tsx';
import {Step} from '@/dialogs/StepsDialog/Step.ts';
import cls from '@/dialogs/StepsDialog/StepsDialog.module.css';

export interface StepsDialogProps<TContext> {
    steps: Step<TContext>[];
    initialContext: TContext;
    icon?: string;
    eyebrow?: string;
    eyebrowDetail?: string;
    onFinish?(context: TContext): void;
    onCancel?(): void;
}

export default function StepsDialog<TContext extends object>(
    {steps, initialContext, icon, eyebrow, eyebrowDetail, onFinish, onCancel}: StepsDialogProps<TContext>) {
    const {CloseDialog} = useDialog();

    const [stepIndex, setStepIndex] = useState(0);
    const [context, setContext] = useState<TContext>(initialContext);

    const currentStep = steps[stepIndex];
    const isFirstStep = stepIndex == 0;
    const isLastStep = stepIndex == steps.length - 1;
    const canProceed = currentStep.canProceed?.(context) ?? true;

    const StepComponent = currentStep.component;

    function updateContext(partial: Partial<TContext>) {
        setContext((prev) => ({...prev, ...partial}));
    }

    function handleFinish() {
        onFinish?.(context);
        CloseDialog();
    }

    function handleCancel() {
        onCancel?.();
        CloseDialog();
    }

    function handleNext() {
        if (isLastStep) {
            handleFinish();
            return;
        }
        setStepIndex(stepIndex + 1);
    }

    function handleBack() {
        if (isFirstStep) {
            handleCancel();
            return;
        }
        setStepIndex(stepIndex - 1);
    }

    return (
        <div className={cls.StepsDialogContainer}>
            <StepsDialogHeader
                icon={icon}
                eyebrow={eyebrow}
                eyebrowDetail={eyebrowDetail}
                stepLabel={currentStep.label ?? currentStep.name}
                stepIndex={stepIndex}
                totalSteps={steps.length}
                onClose={handleCancel}
            />

            <div className={cls.StepsDialogWrapper}>
                <StepComponent {...context} updateContext={updateContext}/>
            </div>

            <StepsDialogFooter
                isLastStep={isLastStep}
                canProceed={canProceed}
                onBack={handleBack}
                onNext={handleNext}
            />
        </div>
    );
}
