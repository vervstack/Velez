import Button from '@/components/base/Button.tsx';
import cls from '@/dialogs/StepsDialog/StepsDialogFooter.module.css';

interface StepsDialogFooterProps {
    isLastStep: boolean;
    canProceed: boolean;
    onBack(): void;
    onNext(): void;
}

export default function StepsDialogFooter({isLastStep, canProceed, onBack, onNext}: StepsDialogFooterProps) {
    return (
        <div className={cls.StepsDialogFooterContainer}>
            <div className={cls.StepsDialogFooterWrapper}>
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
