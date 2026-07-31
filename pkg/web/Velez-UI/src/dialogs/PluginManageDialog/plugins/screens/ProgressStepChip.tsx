import cn from 'classnames';

import {TaskStatusStatus} from '@/app/api/velez';
import cls from '@/dialogs/PluginManageDialog/plugins/screens/ProgressStepChip.module.css';

interface ProgressStepChipProps {
    index: number;
    name: string;
    status?: TaskStatusStatus;
}

const JOB_LABELS: Record<string, string> = {
    generate_credentials: 'Generating credentials',
    create_container: 'Creating container',
    start_container: 'Starting container',
    wait_for_postgres_ready: 'Waiting for Postgres',
    get_root_dsn: 'Resolving connection',
    create_schema_and_migrate: 'Preparing schema',
    create_pg_user: 'Creating cluster user',
    update_cluster_state: 'Updating cluster state',
    init_node_storage: 'Registering node',
    register_plugin: 'Registering plugin',
};

function humanizeJobName(name: string): string {
    return name
        .split('_')
        .filter(Boolean)
        .map(capitalize)
        .join(' ');
}

function capitalize(part: string): string {
    return part.charAt(0).toUpperCase() + part.slice(1);
}

function jobLabel(name: string): string {
    return JOB_LABELS[name] ?? humanizeJobName(name);
}

function statusModifierClass(status: ProgressStepChipProps['status']): string {
    switch (status) {
        case TaskStatusStatus.DONE:
            return cls.done;
        case TaskStatusStatus.RUNNING:
            return cls.running;
        case TaskStatusStatus.FAILED:
            return cls.failed;
        default:
            return cls.pending;
    }
}

export default function ProgressStepChip({index, name, status}: ProgressStepChipProps) {
    const modifier = statusModifierClass(status);
    const isRunning = status === TaskStatusStatus.RUNNING;

    return (
        <div className={cn(cls.ProgressStepChipContainer, modifier)}>
            <div className={cls.ProgressStepChipWrapper}>
                <span className={cls.Dot}/>
                <span className={cls.Eyebrow}>{String(index).padStart(2, '0')}</span>
                <span className={cls.Label}>{jobLabel(name)}</span>
            </div>
            {isRunning && <span className={cls.Shimmer}/>}
        </div>
    );
}
