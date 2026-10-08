import {useState} from "react"

import cls from "@/dialogs/CreateDindDialog/CreateDindDialog.module.css"
import type {CreateDindRequest, CreateDindResponse} from "@/app/api/velez/dind_api.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {queryClient} from "@/app/queryClient.ts"
import {CreateDindMutation, DINDS_QUERY_KEY} from "@/processes/queries/dinds.ts"
import DialogShell from "@/components/DialogShell/DialogShell.tsx"
import CreateDindDialogForm
    from "@/dialogs/CreateDindDialog/components/CreateDindDialogForm/CreateDindDialogForm.tsx"
import TaskProgressScreen from "@/widgets/TaskProgressScreen/TaskProgressScreen.tsx"

export default function CreateDindDialog() {
    const [submittedReq, setSubmittedReq] = useState<CreateDindRequest | null>(null)

    const {CloseDialog} = useDialog()
    const toaster = useToaster()
    const createDind = CreateDindMutation()

    function handleStart(): Promise<CreateDindResponse> {
        if (!submittedReq) {
            return Promise.reject(new Error("no pending create request"))
        }
        return createDind.mutateAsync(submittedReq)
    }

    function handleSuccess() {
        queryClient.invalidateQueries({queryKey: DINDS_QUERY_KEY})
        queryClient.invalidateQueries({queryKey: ["services"]})
        toaster.bake({title: "Docker daemon created", description: submittedReq?.name ?? "", level: "Info"})
    }

    function renderBody() {
        if (submittedReq) {
            return (
                <TaskProgressScreen
                    title="Creating Docker daemon"
                    metaLine={submittedReq.name ?? ""}
                    start={handleStart}
                    onSuccess={handleSuccess}
                    onClose={CloseDialog}
                />
            )
        }
        return <CreateDindDialogForm onSubmit={setSubmittedReq} onCancel={CloseDialog}/>
    }

    return (
        <div className={cls.CreateDindDialogContainer}>
            <DialogShell
                title="Create Docker daemon"
                onClose={submittedReq ? undefined : CloseDialog}
                isFlush={submittedReq !== null}
            >
                {renderBody()}
            </DialogShell>
        </div>
    )
}
