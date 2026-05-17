import cn from 'classnames';

import cls from '@/components/base/StatusDot.module.css';

type DotStatus = 'running' | 'healthy' | 'degraded' | 'stopped' | 'online' | 'offline' | 'error' | 'creating' | 'pending' | 'enabled' | 'disabled';

interface StatusDotProps {
    status: DotStatus;
    pulse?: boolean;
}

const PULSE_STATUSES: DotStatus[] = ['running', 'healthy'];

export default function StatusDot({status, pulse}: StatusDotProps) {
    const shouldPulse = pulse !== false && PULSE_STATUSES.includes(status);
    return (
        <span
            className={cn(
                cls.StatusDotContainer,
                cls[status],
                {[cls.pulse]: shouldPulse}
            )}
        />
    );
}
