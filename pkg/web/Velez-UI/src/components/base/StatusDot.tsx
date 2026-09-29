import cn from 'classnames';

import cls from '@/components/base/StatusDot.module.css';

export type DotStatus = 'running' | 'healthy' | 'degraded' | 'stopped' | 'online' | 'offline' | 'error' | 'creating' | 'pending' | 'enabled' | 'disabled';

interface StatusDotProps {
    status: DotStatus;
    pulse?: boolean;
    tooltip?: string;
}

const PULSE_STATUSES: DotStatus[] = ['running', 'healthy'];

export default function StatusDot({status, pulse, tooltip}: StatusDotProps) {
    const shouldPulse = pulse !== false && PULSE_STATUSES.includes(status);
    return (
        <span
            className={cn(
                cls.StatusDotContainer,
                cls[status],
                {[cls.pulse]: shouldPulse}
            )}
            data-tooltip-id={tooltip ? 'root-tooltip' : undefined}
            data-tooltip-content={tooltip}
            data-tooltip-place="top"
        />
    );
}
