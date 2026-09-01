import {useState} from 'react';

import cls from '@/dialogs/RegistryDeleteDialog/RegistryDeleteDialog.module.css';

import {useToaster} from '@/app/hooks/toaster/Toaster.ts';
import {DeleteRegistryMutation} from '@/processes/queries/control_plane.ts';
import Button from '@/components/base/Button.tsx';
import type {Registry} from '@/app/api/velez';

interface RegistryDeleteDialogProps {
    registry: Registry;
    onCancel: () => void;
    onDeleted: () => void;
}

export default function RegistryDeleteDialog({registry, onCancel, onDeleted}: RegistryDeleteDialogProps) {
    const [isRemoving, setIsRemoving] = useState(false);
    const toaster = useToaster();

    const deleteRegistry = DeleteRegistryMutation();

    function handleConfirm() {
        setIsRemoving(true);
        deleteRegistry.mutate(registry.id || '', {
            onSuccess: () => {
                toaster.bake({
                    title: 'Registry removed',
                    description: registry.name || '',
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
                <h2 className={cls.ModalTitle}>Remove registry</h2>
            </div>

            <div className={cls.ModalContent}>
                <p className={cls.ConfirmText}>
                    Are you sure you want to remove <strong>{registry.name}</strong>? This action cannot be
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
