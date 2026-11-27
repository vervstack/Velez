import {useEffect} from "react"
import {useNavigate} from "react-router-dom"
import {Button} from "@vervstack/chures"

import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {Routes} from "@/app/router/Routes.ts"
import {openCreateVcnNetworkDialog} from "@/dialogs/CreateVcnNetworkDialog/CreateVcnNetworkDialog.tsx"
import type {VervPlugin} from "@/model/services/VervPlugins.tsx"
import {serviceService} from "@/processes/api/service.ts"
import cls from "@/dialogs/PluginManageDialog/PluginManageDialog.module.css"

export default function HeadscalePluginForm({state, serviceName}: VervPlugin) {
    const closeDialog = useDialog((dialog) => dialog.CloseDialog)
    const catchGrpc = useToaster((toaster) => toaster.catchGrpc)
    const navigate = useNavigate()

    useEffect(() => {
        if (state !== "running") return

        navigate(Routes.Service + "/" + serviceName)
        closeDialog()
    }, [state, serviceName, navigate, closeDialog])

    function handleRestart() {
        serviceService.restartService(serviceName).catch(catchGrpc)
    }

    if (state === "running") return null

    if (state === "dead") {
        return (
            <div className={cls.ActionSection}>
                <span>Service is down. Run it?</span>
                <Button type="button" onClick={handleRestart}>Restart</Button>
            </div>
        )
    }

    return (
        <div className={cls.ActionSection}>
            <Button type="button" onClick={openCreateVcnNetworkDialog}>Create network</Button>
        </div>
    )
}
