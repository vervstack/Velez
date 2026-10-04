import cls from "@/dialogs/CreateDindDialog/CreateDindDialog.module.css"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import Button from "@/components/base/Button.tsx"
import CreateDindForm from "@/widgets/CreateDindForm/CreateDindForm.tsx"

export default function CreateDindDialog() {
    const {CloseDialog} = useDialog()

    return (
        <div className={cls.CreateDindDialogContainer}>
            <div className={cls.Header}>
                <h2 className={cls.Title}>Create Docker daemon</h2>
                <Button variant="ghost" sm onClick={CloseDialog}>✕</Button>
            </div>
            <CreateDindForm onCreated={CloseDialog} onCancel={CloseDialog}/>
        </div>
    )
}
