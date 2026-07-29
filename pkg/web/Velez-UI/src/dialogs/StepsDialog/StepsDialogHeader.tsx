import cls from '@/dialogs/StepsDialog/StepsDialogHeader.module.css';

interface StepsDialogHeaderProps {
    icon?: string;
    eyebrow?: string;
    eyebrowDetail?: string;
    stepLabel: string;
    stepIndex: number;
    totalSteps: number;
    onClose(): void;
}

export default function StepsDialogHeader(
    {icon = '≡', eyebrow, eyebrowDetail, stepLabel, stepIndex, totalSteps, onClose}: StepsDialogHeaderProps) {
    return (
        <div className={cls.StepsDialogHeaderContainer}>
            <div className={cls.IconChip}>{icon}</div>

            <div className={cls.TitleWrapper}>
                {eyebrow && (
                    <div className={cls.EyebrowRow}>
                        <span className={cls.Eyebrow}>{eyebrow}</span>
                        {eyebrowDetail && (
                            <>
                                <span className={cls.EyebrowSeparator}>/</span>
                                <span className={cls.EyebrowDetail}>{eyebrowDetail}</span>
                            </>
                        )}
                    </div>
                )}

                <div className={cls.StepLine}>
                    Step {stepIndex + 1}/{totalSteps} · {stepLabel}
                </div>
            </div>

            <div className={cls.Spacer}/>

            <button type="button" className={cls.CloseButton} onClick={onClose}>
                ✕
            </button>
        </div>
    );
}
