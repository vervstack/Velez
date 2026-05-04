import cls from '@/components/base/SkeletonLoader.module.css';
import cn from 'classnames';

interface SkeletonLoaderProps {
    shape: 'line' | 'card' | 'block';
    width?: string;
    height?: string;
}

export default function SkeletonLoader({ shape, width, height }: SkeletonLoaderProps) {
    return (
        <div
            className={cn(cls.SkeletonContainer, cls[shape])}
            style={{ width, height }}
        >
            <div className={cls.skeleton} />
        </div>
    );
}