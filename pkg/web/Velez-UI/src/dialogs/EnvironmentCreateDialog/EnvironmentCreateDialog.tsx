import {useState} from 'react';

import cls from '@/dialogs/EnvironmentCreateDialog/EnvironmentCreateDialog.module.css';

import {useToaster} from '@/app/hooks/toaster/Toaster.ts';
import {CreateEnvironmentMutation} from '@/processes/queries/control_plane.ts';
import Button from '@/components/base/Button.tsx';
import Input from '@/components/base/Input.tsx';

interface EnvironmentCreateDialogProps {
    onCancel: () => void;
    onCreated: () => void;
}

export default function EnvironmentCreateDialog({onCancel, onCreated}: EnvironmentCreateDialogProps) {
    const [name, setName] = useState('');
    const [suffix, setSuffix] = useState('');
    const toaster = useToaster();

    const createEnvironment = CreateEnvironmentMutation();

    function handleCreate() {
        const trimmedName = name.trim();
        if (!trimmedName) {
            return;
        }

        createEnvironment.mutate({name: trimmedName, suffix: suffix.trim() || undefined}, {
            onSuccess: () => {
                toaster.bake({
                    title: 'Environment created',
                    description: trimmedName,
                    level: 'Info',
                });
                onCreated();
            },
            onError: toaster.catchGrpc,
        });
    }

    return (
        <div className={cls.ModalContainer}>
            <div className={cls.ModalHeader}>
                <h2 className={cls.ModalTitle}>Add environment</h2>
            </div>

            <div className={cls.ModalContent}>
                <div className={cls.FieldsWrapper}>
                    <Input
                        label="Name"
                        inputValue={name}
                        onChange={setName}
                        disabled={createEnvironment.isPending}
                    />
                    <Input
                        label="Suffix (optional)"
                        inputValue={suffix}
                        onChange={setSuffix}
                        disabled={createEnvironment.isPending}
                    />
                </div>

                <div className={cls.ActionsRow}>
                    <Button variant="secondary" onClick={onCancel} disabled={createEnvironment.isPending}>
                        Cancel
                    </Button>
                    <Button
                        variant="primary"
                        onClick={handleCreate}
                        disabled={createEnvironment.isPending || !name.trim()}
                    >
                        {createEnvironment.isPending ? 'Creating…' : 'Create'}
                    </Button>
                </div>
            </div>
        </div>
    );
}
