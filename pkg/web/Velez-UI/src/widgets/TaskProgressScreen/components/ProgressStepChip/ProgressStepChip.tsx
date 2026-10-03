import cn from 'classnames';

import {TaskStatusStatus} from '@/app/api/velez';
import cls from '@/widgets/TaskProgressScreen/components/ProgressStepChip/ProgressStepChip.module.css';

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
    put_secret: 'Storing credentials',
    resolve_ports: 'Resolving ports',
    drop_container: 'Dropping container',
    write_htpasswd: 'Writing auth file',
    deploy_registry: 'Deploying registry',
    deploy_registry_ui: 'Deploying UI',
    register_registry_instance_row: 'Registering instance',
    register_registry_row: 'Registering registry',
    bind_owner_resource: 'Binding owner service',
    inspect_container: 'Inspecting container',
    verify_pg_login: 'Verifying Postgres login',
    store_pg_secret: 'Storing Postgres password',
    store_runner_secrets: 'Storing runner secrets',
    verify_registry_login: 'Verifying registry login',
    store_registry_secret: 'Storing registry password',
    link_bind_mounts: 'Linking bind mounts',
    upsert_service: 'Registering service',
    bind_existing_container: 'Binding container',
    bind_sidecars: 'Binding sidecars',
    upsert_runner_row: 'Registering runner',
    upsert_registry_row: 'Registering registry',
    upsert_pg_instance: 'Registering Postgres instance',
    recreate_with_labels: 'Recreating container',
    recreate_sidecars: 'Recreating sidecars',
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
