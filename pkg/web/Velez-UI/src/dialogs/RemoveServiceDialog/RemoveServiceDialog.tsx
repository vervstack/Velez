import {useState} from 'react';

import cls from '@/dialogs/RemoveServiceDialog/RemoveServiceDialog.module.css';

import {Toast, useToaster} from '@/app/hooks/toaster/Toaster.ts';
import {serviceService} from '@/processes/api/service.ts';
import Button from '@/components/base/Button.tsx';

interface RemoveServiceDialogProps {
    serviceName: string;
    onCancel: () => void;
    onRemoved: () => void;
}

export default function RemoveServiceDialog({serviceName, onCancel, onRemoved}: RemoveServiceDialogProps) {
    const [dropRunningInstances, setDropRunningInstances] = useState(false);
    const [isRemoving, setIsRemoving] = useState(false);
    const toaster = useToaster();

    function handleDropToggle() {
        setDropRunningInstances(!dropRunningInstances);
    }

    function handleConfirm() {
        setIsRemoving(true);
        serviceService.removeService(serviceName, dropRunningInstances)
            .then(() => {
                toaster.bake({
                    title: 'Service removed',
                    description: serviceName,
                    level: 'Info',
                } as Toast);
                onRemoved();
            })
            .catch(toaster.catchGrpc)
            .finally(() => setIsRemoving(false));
    }

    return (
        <div className={cls.ModalContainer}>
            <div className={cls.ModalHeader}>
                <h2 className={cls.ModalTitle}>Remove service</h2>
            </div>

            <div className={cls.ModalContent}>
                <p className={cls.ConfirmText}>
                    Are you sure you want to remove <strong>{serviceName}</strong>? This action cannot be undone.
                </p>

                <label className={cls.CheckboxLabel}>
                    <input
                        type="checkbox"
                        checked={dropRunningInstances}
                        onChange={handleDropToggle}
                    />
                    Also drop running instances
                </label>

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
