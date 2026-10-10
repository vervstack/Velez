import {useEffect, useState} from "react"

import cls from "@/dialogs/CreateServiceDialog/screens/DindScreen/DindScreen.module.css"
import type {CreateDindRequest, CreateDindResponse} from "@/app/api/velez/dind_api.pb"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {queryClient} from "@/app/queryClient.ts"
import {CreateDindMutation, DINDS_QUERY_KEY} from "@/processes/queries/dinds.ts"
import CreateDindDialogForm
    from "@/dialogs/CreateServiceDialog/screens/DindScreen/components/CreateDindDialogForm/CreateDindDialogForm.tsx"
import TaskProgressScreen from "@/widgets/TaskProgressScreen/TaskProgressScreen.tsx"

interface Props {
    onBusyChange(isBusy: boolean): void
}

export default function DindScreen({onBusyChange}: Props) {
    const [submittedReq, setSubmittedReq] = useState<CreateDindRequest | null>(null)

    const {CloseDialog} = useDialog()
    const toaster = useToaster()
    const createDind = CreateDindMutation()

    useEffect(() => {
        onBusyChange(submittedReq !== null || createDind.isPending)
    }, [submittedReq, createDind.isPending])

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

    function handleBack() {
        setSubmittedReq(null)
    }

    function renderProgress(req: CreateDindRequest) {
        return (
            <TaskProgressScreen
                title="Creating Docker daemon"
                metaLine={req.name ?? ""}
                start={handleStart}
                onSuccess={handleSuccess}
                onClose={CloseDialog}
                onBack={handleBack}
            />
        )
    }

    return (
        <>
            <div className={cls.DindScreenContainer} hidden={submittedReq !== null}>
                <CreateDindDialogForm onSubmit={setSubmittedReq} onCancel={CloseDialog}/>
            </div>
            {submittedReq && renderProgress(submittedReq)}
        </>
    )
}
