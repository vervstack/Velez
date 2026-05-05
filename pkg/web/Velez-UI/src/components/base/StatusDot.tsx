import cn from 'classnames';

import cls from '@/components/base/StatusDot.module.css';

import {NodeStatus} from "@/app/api/velez";

interface StatusDotProps {
    status: 'running' | 'degraded' | 'stopped' | 'online' | 'offline' | 'enabled' | 'disabled';
    pulse?: boolean;
}

export default function StatusDot({status, pulse}: StatusDotProps) {
    return (
        <span
            className={cn(
                cls.StatusDotContainer,
                cls[status],
                {[cls.pulse]: pulse && status === NodeStatus.NodeStatus_Online}
            )}
        />
    );
}
