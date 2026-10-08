import {useState} from "react"

import cls from "@/dialogs/CreateS3InstanceDialog/CreateS3InstanceDialog.module.css"
import type {CreateS3InstanceRequest, CreateS3InstanceResponse} from "@/app/api/velez/s3_api.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {queryClient} from "@/app/queryClient.ts"
import {CreateS3InstanceMutation, S3_INSTANCES_QUERY_KEY} from "@/processes/queries/s3.ts"
import DialogShell from "@/components/DialogShell/DialogShell.tsx"
import CreateS3InstanceForm
    from "@/dialogs/CreateS3InstanceDialog/components/CreateS3InstanceForm/CreateS3InstanceForm.tsx"
import TaskProgressScreen from "@/widgets/TaskProgressScreen/TaskProgressScreen.tsx"

export default function CreateS3InstanceDialog() {
    const [submittedReq, setSubmittedReq] = useState<CreateS3InstanceRequest | null>(null)

    const {CloseDialog} = useDialog()
    const toaster = useToaster()
    const createInstance = CreateS3InstanceMutation()

    function handleStart(): Promise<CreateS3InstanceResponse> {
        if (!submittedReq) {
            return Promise.reject(new Error("no pending create request"))
        }
        return createInstance.mutateAsync(submittedReq)
    }

    function handleSuccess() {
        queryClient.invalidateQueries({queryKey: S3_INSTANCES_QUERY_KEY})
        queryClient.invalidateQueries({queryKey: ["services"]})
        toaster.bake({title: "S3 instance created", description: submittedReq?.name ?? "", level: "Info"})
    }

    function handleBack() {
        setSubmittedReq(null)
    }

    function renderProgress(req: CreateS3InstanceRequest) {
        return (
            <TaskProgressScreen
                title="Creating S3 instance"
                metaLine={req.name ?? ""}
                start={handleStart}
                onSuccess={handleSuccess}
                onClose={CloseDialog}
                onBack={handleBack}
            />
        )
    }

    return (
        <div className={cls.CreateS3InstanceDialogContainer}>
            <DialogShell
                title="Create S3 instance"
                onClose={submittedReq ? undefined : CloseDialog}
                isFlush={submittedReq !== null}
            >
                <CreateS3InstanceForm
                    isHidden={submittedReq !== null}
                    onSubmit={setSubmittedReq}
                    onCancel={CloseDialog}
                />
                {submittedReq && renderProgress(submittedReq)}
            </DialogShell>
        </div>
    )
}
