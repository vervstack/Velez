import {useState} from "react"
import {useNavigate} from "react-router-dom"

import cls from "@/dialogs/PluginManageDialog/PluginManageDialog.module.css"
import {EnablePluginResponse, EnableRegistry, VervPluginState} from "@/app/api/velez"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {queryClient} from "@/app/queryClient.ts"
import {Routes} from "@/app/router/Routes.ts"
import Button from "@/components/base/Button.tsx"
import Input from "@/components/base/Input.tsx"
import TaskProgressScreen from "@/dialogs/PluginManageDialog/plugins/screens/TaskProgressScreen.tsx"
import {VervPlugin} from "@/model/services/VervPlugins.tsx"
import {controlPlaneService} from "@/processes/api/control_plane.ts"
import {serviceService} from "@/processes/api/service.ts"

type Phase = "form" | "progress"

export default function RegistryPluginForm(pl: VervPlugin) {
    const toaster = useToaster()
    const {CloseDialog} = useDialog()
    const navigate = useNavigate()

    const [phase, setPhase] = useState<Phase>("form")
    const [exposeToPort, setExposeToPort] = useState("")
    const [username, setUsername] = useState("")

    if (pl.state == VervPluginState.running) {
        navigate(Routes.Service + "/" + pl.serviceName)
        CloseDialog()
    }

    function handleEnable() {
        setPhase("progress")
    }

    function handleRestart() {
        serviceService.restartService(pl.serviceName).catch(toaster.catchGrpc)
    }

    function handleStart(): Promise<EnablePluginResponse> {
        const payload: EnableRegistry = {
            exposeToPort: exposeToPort ? parseInt(exposeToPort, 10) : undefined,
            username: username || undefined,
        }
        return controlPlaneService.enableRegistry(payload)
    }

    function handleSuccess() {
        queryClient.invalidateQueries({queryKey: ["plugins"]})
    }

    if (phase == "progress") {
        return (
            <TaskProgressScreen
                title="Enabling registry"
                metaLine="Container registry"
                start={handleStart}
                onSuccess={handleSuccess}
                onClose={CloseDialog}
            />
        )
    }

    if (pl.state == VervPluginState.dead) {
        return (
            <div className={cls.ActionSection}>
                <span>Service is down. Run it?</span>
                <Button onClick={handleRestart}>
                    Restart
                </Button>
            </div>
        )
    }

    if (pl.state == VervPluginState.disabled) {
        return (
            <div className={cls.ActionSection}>
                <div className={cls.ConfigSection}>
                    <Input
                        label="Exposed port (optional)"
                        inputValue={exposeToPort}
                        onChange={setExposeToPort}
                    />
                    <Input
                        label="Username (default: verv)"
                        inputValue={username}
                        onChange={setUsername}
                    />
                </div>

                <Button onClick={handleEnable}>
                    Enable
                </Button>
            </div>
        )
    }

    return (
        <div className={cls.ActionSection}>
            <Button onClick={handleEnable}>
                Enable
            </Button>
        </div>
    )
}
