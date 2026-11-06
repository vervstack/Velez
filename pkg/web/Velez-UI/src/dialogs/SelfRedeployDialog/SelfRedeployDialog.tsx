import {useState} from 'react';

import cls from '@/dialogs/SelfRedeployDialog/SelfRedeployDialog.module.css';

import {useToaster} from '@/app/hooks/toaster/Toaster.ts';
import {useDialog} from '@/app/hooks/dialog/Dialog.tsx';
import {SelfRedeployMutation} from '@/processes/queries/control_plane.ts';
import Button from '@/components/base/Button.tsx';
import Input from '@/components/base/Input.tsx';

export default function SelfRedeployDialog() {
    const [image, setImage] = useState('');

    const toaster = useToaster();
    const {CloseDialog} = useDialog();

    const selfRedeploy = SelfRedeployMutation();

    function handleRedeploy() {
        selfRedeploy.mutateAsync(image.trim())
            .then((response) => {
                if (response.accepted) {
                    toaster.bake({
                        title: 'Redeploy started',
                        description: 'Velez is redeploying itself. Track progress and errors in the velez-swapper container logs.',
                        level: 'Info',
                    });
                }
                CloseDialog();
            })
            .catch(toaster.catchGrpc);
    }

    return (
        <div className={cls.ModalContainer}>
            <div className={cls.ModalHeader}>
                <h2 className={cls.ModalTitle}>Self redeploy</h2>
            </div>

            <div className={cls.ModalContent}>
                <p className={cls.ConfirmText}>
                    Velez will replace its own container. Leave the image empty to re-pull the
                    current image, or specify a different one.
                </p>

                <div className={cls.FieldsWrapper}>
                    <Input
                        label="Image"
                        inputValue={image}
                        onChange={setImage}
                        placeholder="empty = current image, re-pulled"
                        disabled={selfRedeploy.isPending}
                    />
                </div>

                <div className={cls.ActionsRow}>
                    <Button variant="secondary" onClick={CloseDialog} disabled={selfRedeploy.isPending}>
                        Cancel
                    </Button>
                    <Button
                        variant="primary"
                        onClick={handleRedeploy}
                        disabled={selfRedeploy.isPending}
                    >
                        {selfRedeploy.isPending ? 'Redeploying…' : 'Redeploy'}
                    </Button>
                </div>
            </div>
        </div>
    );
}
