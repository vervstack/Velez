import {useState} from 'react';

import cls from '@/dialogs/RegistryCreateDialog/RegistryCreateDialog.module.css';

import {useToaster} from '@/app/hooks/toaster/Toaster.ts';
import {CreateRegistryMutation} from '@/processes/queries/control_plane.ts';
import {RegistryType} from '@/app/api/velez';
import Button from '@/components/base/Button.tsx';
import Input from '@/components/base/Input.tsx';
import Choice from '@/components/base/Choice.tsx';
import Checkbox from '@/components/base/Checkbox.tsx';

interface RegistryCreateDialogProps {
    onCancel: () => void;
    onCreated: () => void;
}

export default function RegistryCreateDialog({onCancel, onCreated}: RegistryCreateDialogProps) {
    const [name, setName] = useState('');
    const [type, setType] = useState<RegistryType>(RegistryType.REGISTRY_TYPE_DOCKERHUB);
    const [url, setUrl] = useState('');
    const [username, setUsername] = useState('');
    const [secret, setSecret] = useState('');
    const [isDefault, setIsDefault] = useState(false);
    const toaster = useToaster();

    const createRegistry = CreateRegistryMutation();

    const isGeneric = type === RegistryType.REGISTRY_TYPE_GENERIC_V2;

    function handleCreate() {
        const trimmedName = name.trim();
        if (!trimmedName) {
            return;
        }

        createRegistry.mutate({
            name: trimmedName,
            type,
            url: isGeneric ? (url.trim() || undefined) : undefined,
            username: isGeneric ? (username.trim() || undefined) : undefined,
            secret: isGeneric ? (secret.trim() || undefined) : undefined,
            isDefault,
        }, {
            onSuccess: () => {
                toaster.bake({
                    title: 'Registry created',
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
                <h2 className={cls.ModalTitle}>Add registry</h2>
            </div>

            <div className={cls.ModalContent}>
                <div className={cls.FieldsWrapper}>
                    <Input
                        label="Name"
                        inputValue={name}
                        onChange={setName}
                        disabled={createRegistry.isPending}
                    />

                    <div className={cls.ChoiceRow}>
                        <Choice
                            title="Docker Hub"
                            active={type === RegistryType.REGISTRY_TYPE_DOCKERHUB}
                            disabled={createRegistry.isPending}
                            onClick={() => setType(RegistryType.REGISTRY_TYPE_DOCKERHUB)}
                        />
                        <Choice
                            title="Private registry (v2)"
                            active={isGeneric}
                            disabled={createRegistry.isPending}
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
                            disabled={createRegistry.isPending}
                        />
                    )}

                    <Checkbox
                        label="Set as default"
                        checked={isDefault}
                        onChange={setIsDefault}
                    />
                </div>

                <div className={cls.ActionsRow}>
                    <Button variant="secondary" onClick={onCancel} disabled={createRegistry.isPending}>
                        Cancel
                    </Button>
                    <Button
                        variant="primary"
                        onClick={handleCreate}
                        disabled={createRegistry.isPending || !name.trim()}
                    >
                        {createRegistry.isPending ? 'Creating…' : 'Create'}
                    </Button>
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
            <Input label="Secret" inputValue={secret} onChange={onSecretChange} disabled={disabled}/>
        </div>
    );
}
