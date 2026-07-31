import {useEffect, useState} from 'react';
import cn from 'classnames';

import {EnableStatefullCluster, TaskStatus, TaskStatusStatus, WatchTaskRequest} from '@/app/api/velez';
import {useToaster} from '@/app/hooks/toaster/Toaster.ts';
import Button from '@/components/base/Button.tsx';
import {StatefullPgContext} from '@/dialogs/PluginManageDialog/plugins/StatefullPgContext.ts';
import ProgressStepChip from '@/dialogs/PluginManageDialog/plugins/screens/ProgressStepChip.tsx';
import cls from '@/dialogs/PluginManageDialog/plugins/screens/StatefullPgProgressScreen.module.css';
import {controlPlaneService} from '@/processes/api/control_plane.ts';
import {WatchTaskStream} from '@/processes/api/tasks.ts';
import {NodeHardwareQuery} from '@/processes/queries/control_plane.ts';

interface StatefullPgProgressScreenProps {
    context: StatefullPgContext;
    onClose(): void;
}

type ProgressPhase = 'pending' | 'running' | 'done' | 'failed';

function phaseEyebrow(phase: ProgressPhase): string {
    if (phase === 'done') {
        return 'success';
    }
    if (phase === 'failed') {
        return 'failed';
    }
    return 'deploying';
}

function headerTitle(taskId: string | undefined): string {
    return taskId ? `Enabling cluster mode · Task #${taskId}` : 'Enabling cluster mode…';
}

function eyebrowModifierClass(phase: ProgressPhase): string {
    if (phase === 'done') {
        return cls.done;
    }
    if (phase === 'failed') {
        return cls.failed;
    }
    return '';
}

export default function StatefullPgProgressScreen({context, onClose}: StatefullPgProgressScreenProps) {
    const [phase, setPhase] = useState<ProgressPhase>('pending');
    const [error, setError] = useState<string | undefined>(undefined);
    const [taskStatus, setTaskStatus] = useState<TaskStatus | undefined>(undefined);

    const hardwareQuery = NodeHardwareQuery();
    const nodeRegion = hardwareQuery.data?.nodeRegion;

    useEffect(() => {
        let cancelled = false;

        let finalStatus: TaskStatusStatus | undefined;
        let finalError: string | undefined;

        function onStatus(status: TaskStatus) {
            if (cancelled) {
                return;
            }
            finalStatus = status.status;
            finalError = status.error;
            setTaskStatus(status);
            if (status.status === TaskStatusStatus.RUNNING) {
                setPhase('running');
            }
        }

        const payload: EnableStatefullCluster = {
            isExposePort: context.exposePort,
            exposeToPort: context.exposePort ? context.portNumber : undefined,
        };

        controlPlaneService.enableStatefullPgCluster(payload)
            .then((res) => {
                const watchReq: WatchTaskRequest = {entityId: res.entityId, action: res.action};
                return WatchTaskStream(watchReq, onStatus);
            })
            .then(() => {
                if (cancelled) {
                    return;
                }
                if (finalStatus === TaskStatusStatus.FAILED) {
                    throw new Error(finalError || 'Enabling cluster mode failed');
                }
                setPhase('done');
            })
            .catch((err: Error) => {
                if (cancelled) {
                    return;
                }
                setPhase('failed');
                setError(err.message);
                useToaster.getState().catchGrpc(err);
            });

        return () => {
            cancelled = true;
        };
    }, []);

    const jobs = taskStatus?.jobs;
    const hasSteps = jobs !== undefined && jobs.length > 0;

    return (
        <div className={cls.ProgressContainer}>
            <div className={cls.HeaderWrapper}>
                {(phase === 'pending' || phase === 'running') && (
                    <span className={cn(cls.Glyph, cls.Spinner)}/>
                )}
                {phase === 'done' && <span className={cn(cls.Glyph, cls.SuccessMark)}>✓</span>}
                {phase === 'failed' && <span className={cn(cls.Glyph, cls.FailureMark)}>!</span>}

                <span className={cn(cls.Eyebrow, eyebrowModifierClass(phase))}>{phaseEyebrow(phase)}</span>
                <span className={cls.Title}>{headerTitle(taskStatus?.taskId)}</span>

                <span className={cls.MetaLine}>PostgreSQL cluster</span>
                {nodeRegion && <span className={cn(cls.MetaLine, cls.MetaLineSecondary)}>{nodeRegion}</span>}
            </div>

            {hasSteps && (
                <div className={cls.StepsWrapper}>
                    {jobs.map((job, idx) => (
                        <ProgressStepChip
                            key={job.name ?? idx}
                            index={idx + 1}
                            name={job.name ?? ''}
                            status={job.status}
                        />
                    ))}
                </div>
            )}

            {(phase === 'done' || phase === 'failed') && (
                <div className={cls.ActionsWrapper}>
                    {phase === 'failed' && error && <span className={cls.ErrorText}>{error}</span>}
                    {phase === 'done' && <Button variant="primary" onClick={onClose}>Done</Button>}
                    {phase === 'failed' && <Button variant="secondary" onClick={onClose}>Close</Button>}
                </div>
            )}
        </div>
    );
}
