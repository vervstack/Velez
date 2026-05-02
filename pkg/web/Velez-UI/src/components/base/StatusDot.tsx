import cls from '@/components/base/StatusDot.module.css';
import cn from 'classnames';

interface StatusDotProps {
    status: 'running' | 'degraded' | 'stopped' | 'online' | 'offline';
    pulse?: boolean;
}

export default function StatusDot({ status, pulse }: StatusDotProps) {
    return (
        <span
            className={cn(
                cls.StatusDotContainer,
                cls[status],
                { [cls.pulse]: pulse && status === 'running' }
            )}
        />
    );
}
