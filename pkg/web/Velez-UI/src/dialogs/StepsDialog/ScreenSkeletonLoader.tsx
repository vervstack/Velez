import cls from '@/dialogs/StepsDialog/ScreenSkeletonLoader.module.css';

export default function ScreenSkeletonLoader() {
    return (
        <div className={cls.ScreenSkeletonLoaderContainer} data-testid="screen-skeleton-loader">
            <div className={cls.TitleBar}/>

            <div className={cls.BodyLineWrapper}>
                <div className={cls.BodyLine}/>
                <div className={cls.BodyLine}/>
                <div className={cls.BodyLineShort}/>
            </div>
        </div>
    );
}
