import cls from '@/dialogs/StepsDialog/ScreenSkeletonLoader.module.css';

interface ScreenSkeletonLoaderProps {
    hint?: string;
}

export default function ScreenSkeletonLoader({hint}: ScreenSkeletonLoaderProps) {
    return (
        <div className={cls.ScreenSkeletonLoaderContainer} data-testid="screen-skeleton-loader">
            <div className={cls.TitleBar}/>

            <div className={cls.BodyLineWrapper}>
                <div className={cls.BodyLine}/>
                <div className={cls.BodyLine}/>
                <div className={cls.BodyLineShort}/>
            </div>

            {hint && (
                <p className={cls.Hint}>{hint}</p>
            )}
        </div>
    );
}
