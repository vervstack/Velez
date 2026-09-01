import {useState} from 'react';

import cls from '@/dialogs/RegistryManageDialog/RegistryManageDialog.module.css';

import {useToaster} from '@/app/hooks/toaster/Toaster.ts';
import {useDialog} from '@/app/hooks/dialog/Dialog.tsx';
import {UpdateRegistryMutation} from '@/processes/queries/control_plane.ts';
import {RegistryType} from '@/app/api/velez';
import Button from '@/components/base/Button.tsx';
import Input from '@/components/base/Input.tsx';
import Choice from '@/components/base/Choice.tsx';
import Checkbox from '@/components/base/Checkbox.tsx';
import RegistryDeleteDialog from '@/dialogs/RegistryDeleteDialog/RegistryDeleteDialog.tsx';
import type {Registry} from '@/app/api/velez';

interface RegistryManageDialogProps {
    registry: Registry;
    onCancel: () => void;
    onSaved: () => void;
    onDeleted: () => void;
}

export default function RegistryManageDialog({registry, onCancel, onSaved, onDeleted}: RegistryManageDialogProps) {
    const [name, setName] = useState(registry.name || '');
    const [type, setType] = useState<RegistryType>(registry.type || RegistryType.REGISTRY_TYPE_DOCKERHUB);
    const [url, setUrl] = useState(registry.url || '');
    const [username, setUsername] = useState(registry.username || '');
    const [secret, setSecret] = useState('');
    const [isDefault, setIsDefault] = useState(registry.isDefault || false);
    const toaster = useToaster();
    const {OpenDialog, CloseDialog} = useDialog();

    const updateRegistry = UpdateRegistryMutation();

    const isGeneric = type === RegistryType.REGISTRY_TYPE_GENERIC_V2;

    function handleSave() {
        const trimmedName = name.trim();
        if (!trimmedName) {
            return;
        }

        updateRegistry.mutate({
            id: registry.id || '',
            name: trimmedName,
            type,
            url: isGeneric ? (url.trim() || undefined) : undefined,
            username: isGeneric ? (username.trim() || undefined) : undefined,
            secret: isGeneric ? (secret.trim() || undefined) : undefined,
            isDefault,
        }, {
            onSuccess: () => {
                toaster.bake({
                    title: 'Registry updated',
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
            <RegistryDeleteDialog
                registry={registry}
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
                <h2 className={cls.ModalTitle}>Manage registry</h2>
            </div>

            <div className={cls.ModalContent}>
                <div className={cls.FieldsWrapper}>
                    <Input
                        label="Name"
                        inputValue={name}
                        onChange={setName}
                        disabled={updateRegistry.isPending}
                    />

                    <div className={cls.ChoiceRow}>
                        <Choice
                            title="Docker Hub"
                            active={type === RegistryType.REGISTRY_TYPE_DOCKERHUB}
                            disabled={updateRegistry.isPending}
                            onClick={() => setType(RegistryType.REGISTRY_TYPE_DOCKERHUB)}
                        />
                        <Choice
                            title="Private registry (v2)"
                            active={isGeneric}
                            disabled={updateRegistry.isPending}
                            onClick={() => setType(RegistryType.REGISTRY_TYPE_GENERIC_V2)}
                        />
                    </div>

                    {isGeneric && (
                        <RegistryConnectionFields
                            url={url}
                            username={username}
                            secret={secret}
                            onUrlChange={setUrl}
                            onUsernameChange={setUsername}
                            onSecretChange={setSecret}
                            disabled={updateRegistry.isPending}
                        />
                    )}

                    <Checkbox
                        label="Set as default"
                        checked={isDefault}
                        onChange={setIsDefault}
                    />
                </div>

                <div className={cls.ActionsRow}>
                    <Button
                        variant="danger"
                        borderless
                        onClick={handleDelete}
                        disabled={updateRegistry.isPending}
                    >
                        Delete
                    </Button>

                    <div className={cls.PrimaryActionsWrapper}>
                        <Button variant="secondary" onClick={onCancel} disabled={updateRegistry.isPending}>
                            Cancel
                        </Button>
                        <Button
                            variant="primary"
                            onClick={handleSave}
                            disabled={updateRegistry.isPending || !name.trim()}
                        >
                            {updateRegistry.isPending ? 'Saving…' : 'Save'}
                        </Button>
                    </div>
                </div>
            </div>
        </div>
    );
}

interface RegistryConnectionFieldsProps {
    url: string;
    username: string;
    secret: string;
    onUrlChange: (v: string) => void;
    onUsernameChange: (v: string) => void;
    onSecretChange: (v: string) => void;
    disabled?: boolean;
}

function RegistryConnectionFields(
    {url, username, secret, onUrlChange, onUsernameChange, onSecretChange, disabled}: RegistryConnectionFieldsProps
) {
    return (
        <div className={cls.ConnectionFieldsWrapper}>
            <Input label="Url" inputValue={url} onChange={onUrlChange} disabled={disabled}/>
            <Input label="Username" inputValue={username} onChange={onUsernameChange} disabled={disabled}/>
            <Input
                label="Secret"
                inputValue={secret}
                onChange={onSecretChange}
                disabled={disabled}
                hint="Leave blank to keep current secret"
            />
        </div>
    );
}
