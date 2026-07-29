import Button from '@/components/base/Button.tsx';
import cls from '@/dialogs/StepsDialog/StepsDialogFooter.module.css';

interface StepsDialogFooterProps {
    isLastStep: boolean;
    canProceed: boolean;
    isLoading?: boolean;
    loadingHint?: string;
    onBack(): void;
    onNext(): void;
}

export default function StepsDialogFooter(
    {isLastStep, canProceed, isLoading, loadingHint, onBack, onNext}: StepsDialogFooterProps) {
    return (
        <div className={cls.StepsDialogFooterContainer}>
            <div className={cls.StepsDialogFooterWrapper}>
                {isLoading && loadingHint && (
                    <span className={cls.LoadingHint}>{loadingHint}</span>
                )}

                <Button variant="ghost" onClick={onBack}>
                    Back
                </Button>

                <Button variant="primary" onClick={onNext} disabled={!canProceed}>
                    {isLastStep ? 'Finish' : 'Next'}
                </Button>
            </div>
        </div>
    );
}
