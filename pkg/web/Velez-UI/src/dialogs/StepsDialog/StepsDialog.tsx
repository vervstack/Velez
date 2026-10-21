import {ComponentType, useEffect, useState} from 'react';

import {useDialog} from '@/app/hooks/dialog/Dialog.tsx';
import StepsDialogHeader, {StepsDialogHeaderContent} from '@/dialogs/StepsDialog/StepsDialogHeader.tsx';
import StepsDialogFooter from '@/dialogs/StepsDialog/StepsDialogFooter.tsx';
import StepsDialogStepper from '@/dialogs/StepsDialog/StepsDialogStepper.tsx';
import {ConfirmFinishProps, FinalStepProps, Step} from '@/dialogs/StepsDialog/Step.ts';
import cls from '@/dialogs/StepsDialog/StepsDialog.module.css';

export interface StepsDialogProps<TContext> {
    steps: Step<TContext>[];
    initialContext: TContext;
    header?: StepsDialogHeaderContent;
    isLoading?: boolean;
    loadingHint?: string;
    resolveContext?: Promise<Partial<TContext>>;
    onFinish?(context: TContext): void;
    onCancel?(): void;
    finalStep?: ComponentType<FinalStepProps<TContext>>;
    initialStepIndex?: number;
    railLabel?: string;
    confirmFinish?: ComponentType<ConfirmFinishProps<TContext>>;
    shouldConfirmFinish?(context: TContext): boolean;
}

export default function StepsDialog<TContext extends object>(
    {
        steps, initialContext, header, isLoading = false, loadingHint, resolveContext,
        onFinish, onCancel, finalStep, initialStepIndex = 0, railLabel, confirmFinish,
        shouldConfirmFinish,
    }: StepsDialogProps<TContext>) {
    const {CloseDialog} = useDialog();

    const [stepIndex, setStepIndex] = useState(initialStepIndex);
    const [isConfirmingFinish, setIsConfirmingFinish] = useState(false);
    const [context, setContext] = useState<TContext>(initialContext);
    const [isResolvingContext, setIsResolvingContext] = useState(Boolean(resolveContext));
    const [isFinished, setIsFinished] = useState(false);

    useEffect(() => {
        if (!resolveContext) {
            return;
        }

        let cancelled = false;

        resolveContext
            .then((patch) => {
                if (cancelled) {
                    return;
                }
                setContext((prev) => ({...prev, ...patch}));
            })
            .finally(() => {
                if (!cancelled) {
                    setIsResolvingContext(false);
                }
            });

        return () => {
            cancelled = true;
        };
    }, [resolveContext]);

    const effectiveIsLoading = resolveContext ? isResolvingContext : isLoading;

    const currentStep = steps[stepIndex];
    const isFirstStep = stepIndex == 0;
    const isLastStep = stepIndex == steps.length - 1;
    const canProceed = (currentStep.canProceed?.(context) ?? true) && !effectiveIsLoading;

    const StepComponent = currentStep.component;

    function updateContext(partial: Partial<TContext>) {
        setContext((prev) => ({...prev, ...partial}));
    }

    function finish() {
        if (finalStep) {
            setIsFinished(true);
            return;
        }
        onFinish?.(context);
        CloseDialog();
    }

    function handleFinish() {
        if (confirmFinish && (shouldConfirmFinish?.(context) ?? true)) {
            setIsConfirmingFinish(true);
            return;
        }
        finish();
    }

    function handleConfirmFinish() {
        setIsConfirmingFinish(false);
        finish();
    }

    function handleCancelFinish() {
        setIsConfirmingFinish(false);
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

    function handleStepSelect(index: number) {
        if (index < stepIndex && !effectiveIsLoading) {
            setStepIndex(index);
        }
    }

    const FinalStepComponent = finalStep;
    const showFinalStep = isFinished && Boolean(FinalStepComponent);
    const ConfirmFinishComponent = confirmFinish;

    return (
        <div className={cls.StepsDialogContainer}>
            <StepsDialogHeader
                header={header}
                isLoading={effectiveIsLoading}
                stepLabel={currentStep.label ?? currentStep.name}
                stepIndex={stepIndex}
                totalSteps={steps.length}
                onClose={handleCancel}
            />

            {showFinalStep && FinalStepComponent ? (
                <FinalStepComponent context={context} onClose={CloseDialog}/>
            ) : isConfirmingFinish && ConfirmFinishComponent ? (
                <ConfirmFinishComponent
                    context={context}
                    onConfirm={handleConfirmFinish}
                    onCancel={handleCancelFinish}
                />
            ) : (
                <>
                    <div className={cls.StepsDialogBody}>
                        <StepsDialogStepper
                            steps={steps}
                            currentIndex={stepIndex}
                            isLoading={effectiveIsLoading}
                            railLabel={railLabel}
                            onSelect={handleStepSelect}
                        />

                        <div className={cls.StepsDialogWrapper}>
                            <StepComponent {...context} updateContext={updateContext}/>
                        </div>
                    </div>

                    <StepsDialogFooter
                        isLastStep={isLastStep}
                        nextLabel={currentStep.nextLabel}
                        canProceed={canProceed}
                        isLoading={effectiveIsLoading}
                        loadingHint={loadingHint}
                        onBack={handleBack}
                        onNext={handleNext}
                    />
                </>
            )}
        </div>
    );
}
