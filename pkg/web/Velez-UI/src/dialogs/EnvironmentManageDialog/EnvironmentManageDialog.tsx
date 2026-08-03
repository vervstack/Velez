import {useState} from 'react';

import cls from '@/dialogs/EnvironmentManageDialog/EnvironmentManageDialog.module.css';

import {useToaster} from '@/app/hooks/toaster/Toaster.ts';
import {useDialog} from '@/app/hooks/dialog/Dialog.tsx';
import {UpdateEnvironmentMutation} from '@/processes/queries/control_plane.ts';
import Button from '@/components/base/Button.tsx';
import Input from '@/components/base/Input.tsx';
import EnvironmentDeleteDialog from '@/dialogs/EnvironmentDeleteDialog/EnvironmentDeleteDialog.tsx';
import type {Environment} from '@/app/api/velez';

interface EnvironmentManageDialogProps {
    environment: Environment;
    onCancel: () => void;
    onSaved: () => void;
    onDeleted: () => void;
}

export default function EnvironmentManageDialog({environment, onCancel, onSaved, onDeleted}: EnvironmentManageDialogProps) {
    const [name, setName] = useState(environment.name || '');
    const [suffix, setSuffix] = useState(environment.suffix || '');
    const toaster = useToaster();
    const {OpenDialog, CloseDialog} = useDialog();

    const updateEnvironment = UpdateEnvironmentMutation();

    function handleSave() {
        const trimmedName = name.trim();
        if (!trimmedName) {
            return;
        }

        updateEnvironment.mutate({id: environment.id || '', name: trimmedName, suffix: suffix.trim() || undefined}, {
            onSuccess: () => {
                toaster.bake({
                    title: 'Environment updated',
                    description: trimmedName,
                    level: 'Info',
                });
                onSaved();
            },
            onError: toaster.catchGrpc,
        });
    }

    function handleDelete() {
        OpenDialog(
            <EnvironmentDeleteDialog
                environment={environment}
                onCancel={CloseDialog}
                onDeleted={() => {
                    CloseDialog();
                    onDeleted();
                }}
            />
        );
    }

    return (
        <div className={cls.ModalContainer}>
            <div className={cls.ModalHeader}>
                <h2 className={cls.ModalTitle}>Manage environment</h2>
            </div>

            <div className={cls.ModalContent}>
                <div className={cls.FieldsWrapper}>
                    <Input
                        label="Name"
                        inputValue={name}
                        onChange={setName}
                        disabled={updateEnvironment.isPending}
                    />
                    <Input
                        label="Suffix (optional)"
                        inputValue={suffix}
                        onChange={setSuffix}
                        disabled={updateEnvironment.isPending}
                    />
                </div>

                <div className={cls.ActionsRow}>
                    <Button
                        variant="danger"
                        borderless
                        onClick={handleDelete}
                        disabled={updateEnvironment.isPending}
                    >
                        Delete
                    </Button>

                    <div className={cls.PrimaryActionsWrapper}>
                        <Button variant="secondary" onClick={onCancel} disabled={updateEnvironment.isPending}>
                            Cancel
                        </Button>
                        <Button
                            variant="primary"
                            onClick={handleSave}
                            disabled={updateEnvironment.isPending || !name.trim()}
                        >
                            {updateEnvironment.isPending ? 'Saving…' : 'Save'}
                        </Button>
                    </div>
                </div>
            </div>
        </div>
    );
}
