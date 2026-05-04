import cn from 'classnames';

import cls from '@/components/base/StatusDot.module.css';

import {NodeStatus} from "@/app/api/velez";

interface StatusDotProps {
    status: NodeStatus;
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
