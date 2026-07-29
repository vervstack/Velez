import cn from 'classnames';

import cls from '@/dialogs/StepsDialog/StepsDialogHeader.module.css';

export interface StepsDialogHeaderContent {
    icon?: string;
    eyebrow?: string;
    eyebrowDetail?: string;
}

interface StepsDialogHeaderProps {
    header?: StepsDialogHeaderContent;
    isLoading: boolean;
    stepLabel: string;
    stepIndex: number;
    totalSteps: number;
    onClose(): void;
}

function warnFieldMissing(fieldName: string) {
    console.error(
        `StepsDialogHeader: "${fieldName}" is undefined while isLoading is false — header fields must be provided once loading completes.`
    );
}

export default function StepsDialogHeader(
    {header, isLoading, stepLabel, stepIndex, totalSteps, onClose}: StepsDialogHeaderProps) {
    const icon = header?.icon;
    const eyebrow = header?.eyebrow;
    const eyebrowDetail = header?.eyebrowDetail;

    const isIconSkeleton = icon === undefined;
    const isEyebrowSkeleton = eyebrow === undefined;
    const isEyebrowDetailSkeleton = eyebrowDetail === undefined;

    if (!isLoading && isIconSkeleton) {
        warnFieldMissing('icon');
    }
    if (!isLoading && isEyebrowSkeleton) {
        warnFieldMissing('eyebrow');
    }
    if (!isLoading && isEyebrowDetailSkeleton) {
        warnFieldMissing('eyebrowDetail');
    }

    return (
        <div className={cls.StepsDialogHeaderContainer}>
            <div className={cn(cls.IconChip, isIconSkeleton && cls.Skeleton)}>
                {!isIconSkeleton && icon}
            </div>

            <div className={cls.TitleWrapper}>
                <div className={cls.EyebrowRow}>
                    {isEyebrowSkeleton ? (
                        <span className={cn(cls.EyebrowPill, cls.Skeleton)}/>
                    ) : (
                        <span className={cls.Eyebrow}>{eyebrow}</span>
                    )}

                    {isEyebrowSkeleton === isEyebrowDetailSkeleton && (
                        <span className={cls.EyebrowSeparator}>/</span>
                    )}

                    {isEyebrowDetailSkeleton ? (
                        <span className={cn(cls.EyebrowPill, cls.Skeleton)}/>
                    ) : (
                        <span className={cls.EyebrowDetail}>{eyebrowDetail}</span>
                    )}
                </div>

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
