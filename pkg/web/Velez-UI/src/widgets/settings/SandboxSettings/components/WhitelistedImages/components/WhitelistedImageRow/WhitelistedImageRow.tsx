import {ConfirmDialog} from "@vervstack/chures"

import cls from "@/widgets/settings/SandboxSettings/components/WhitelistedImages/components/WhitelistedImageRow/WhitelistedImageRow.module.css"
import type {WhitelistedImage} from "@/app/api/velez/settings_api.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {RemoveWhitelistedImageMutation} from "@/processes/queries/settings.ts"
import Badge from "@/components/base/Badge.tsx"
import Button from "@/components/base/Button.tsx"

interface Props {
    image: WhitelistedImage
}

export default function WhitelistedImageRow({image}: Props) {
    const {OpenDialog, CloseDialog} = useDialog()
    const toaster = useToaster()
    const removeImage = RemoveWhitelistedImageMutation()

    const repository = image.repository ?? ""

    function handleConfirmRemove() {
        removeImage.mutateAsync({id: image.id})
            .then(CloseDialog)
            .catch(toaster.catchGrpc)
    }

    function handleRemove() {
        OpenDialog(
            <ConfirmDialog
                danger
                title="Remove image?"
                message={`Remove ${repository} from the whitelist? ` +
                    "Containers from this image can no longer be created with elevated access."}
                confirmLabel="Remove"
                onConfirm={handleConfirmRemove}
                onClose={CloseDialog}
            />
        )
    }

    function renderAction() {
        if (image.isBuiltin) {
            return <Badge label="built-in"/>
        }
        return (
            <Button sm variant="danger" onClick={handleRemove} disabled={removeImage.isPending}>
                Remove
            </Button>
        )
    }

    return (
        <div className={cls.WhitelistedImageRowContainer}>
            <span className={cls.Repository}>{repository}</span>
            {renderAction()}
        </div>
    )
}
