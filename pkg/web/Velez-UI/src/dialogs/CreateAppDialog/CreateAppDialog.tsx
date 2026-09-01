import {useState} from 'react';
import {useNavigate} from 'react-router-dom';
import {useQueryClient} from '@tanstack/react-query';

import cls from '@/dialogs/CreateAppDialog/CreateAppDialog.module.css';

import {useToaster} from '@/app/hooks/toaster/Toaster.ts';
import {useDialog} from '@/app/hooks/dialog/Dialog.tsx';
import useSettings from '@/app/settings/state.ts';
import {Routes} from '@/app/router/Routes.ts';
import {DeploySmerdStream} from '@/processes/api/velez.ts';
import {CreateSmerdReq} from '@/model/smerds/Smerds.ts';
import Button from '@/components/base/Button.tsx';
import Input from '@/components/base/Input.tsx';
import RegistryImagePicker from '@/components/RegistryImagePicker/RegistryImagePicker.tsx';
import {deriveAppName} from '@/dialogs/CreateAppDialog/processes/deriveAppName.ts';

export default function CreateAppDialog() {
    const [gitRepo, setGitRepo] = useState('');
    const [name, setName] = useState('');
    const [nameEdited, setNameEdited] = useState(false);
    const [image, setImage] = useState('');
    const [creating, setCreating] = useState(false);

    const navigate = useNavigate();
    const queryClient = useQueryClient();
    const toaster = useToaster();
    const settings = useSettings();
    const {LockClosing, UnlockClosing, CloseDialog} = useDialog();

    function handleNameChange(v: string) {
        setName(v);
        setNameEdited(true);
    }

    function handleGitRepoChange(v: string) {
        setGitRepo(v);
        if (!nameEdited) {
            const derived = deriveAppName(v);
            if (derived) setName(derived);
        }
    }

    function handleCreate() {
        const trimmedName = name.trim();
        const trimmedImage = image.trim();
        if (!trimmedName || !trimmedImage) return;

        const req = new CreateSmerdReq();
        req.name = trimmedName;
        req.imageName = trimmedImage;
        req.gitRepoUrl = gitRepo.trim();

        let finalError = '';

        function onStatus(status: {error?: string}) {
            finalError = status.error || '';
        }

        setCreating(true);
        LockClosing();

        DeploySmerdStream(req, settings.initReq(), onStatus)
            .then(() => {
                if (finalError) {
                    throw new Error(finalError);
                }
                return queryClient.invalidateQueries({queryKey: ['smerds']});
            })
            .then(() => {
                toaster.bake({title: 'App created', description: trimmedName, level: 'Info'});
                UnlockClosing();
                CloseDialog();
                navigate(Routes.Smerd + '/' + trimmedName);
            })
            .catch(toaster.catchGrpc)
            .finally(() => {
                UnlockClosing();
                setCreating(false);
            });
    }

    return (
        <div className={cls.CreateAppDialogContainer}>
            <div className={cls.Header}>
                <h2 className={cls.Title}>Create app</h2>
            </div>
            <div className={cls.Content}>
                <div className={cls.FieldsWrapper}>
                    <Input
                        label="Git repository"
                        inputValue={gitRepo}
                        onChange={handleGitRepoChange}
                        disabled={creating}
                    />
                    <Input label="Name" inputValue={name} onChange={handleNameChange} disabled={creating}/>
                    <RegistryImagePicker label="Image" value={image} onChange={setImage}/>
                </div>
                <div className={cls.ActionsRow}>
                    <Button variant="secondary" onClick={CloseDialog} disabled={creating}>
                        Cancel
                    </Button>
                    <Button
                        variant="primary"
                        onClick={handleCreate}
                        disabled={creating || !name.trim() || !image.trim()}
                    >
                        {creating ? 'Creating…' : 'Create app'}
                    </Button>
                </div>
            </div>
        </div>
    );
}
