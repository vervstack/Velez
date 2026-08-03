import {useState} from 'react';

import cls from '@/dialogs/EnvironmentDeleteDialog/EnvironmentDeleteDialog.module.css';

import {useToaster} from '@/app/hooks/toaster/Toaster.ts';
import {DeleteEnvironmentMutation} from '@/processes/queries/control_plane.ts';
import Button from '@/components/base/Button.tsx';
import type {Environment} from '@/app/api/velez';

interface EnvironmentDeleteDialogProps {
    environment: Environment;
    onCancel: () => void;
    onDeleted: () => void;
}

export default function EnvironmentDeleteDialog({environment, onCancel, onDeleted}: EnvironmentDeleteDialogProps) {
    const [isRemoving, setIsRemoving] = useState(false);
    const toaster = useToaster();

    const deleteEnvironment = DeleteEnvironmentMutation();

    function handleConfirm() {
        setIsRemoving(true);
        deleteEnvironment.mutate(environment.id || '', {
            onSuccess: () => {
                toaster.bake({
                    title: 'Environment removed',
                    description: environment.name || '',
                    level: 'Info',
                });
                onDeleted();
            },
            onError: (e) => {
                toaster.catchGrpc(e);
                setIsRemoving(false);
            },
        });
    }

    return (
        <div className={cls.ModalContainer}>
            <div className={cls.ModalHeader}>
                <h2 className={cls.ModalTitle}>Remove environment</h2>
            </div>

            <div className={cls.ModalContent}>
                <p className={cls.ConfirmText}>
                    Are you sure you want to remove <strong>{environment.name}</strong>? This action cannot be
                    undone.
                </p>

                <div className={cls.ActionsRow}>
                    <Button variant="secondary" onClick={onCancel} disabled={isRemoving}>
                        Cancel
                    </Button>
                    <Button variant="danger" onClick={handleConfirm} disabled={isRemoving}>
                        {isRemoving ? 'Removing…' : 'Remove'}
                    </Button>
                </div>
            </div>
        </div>
    );
}
