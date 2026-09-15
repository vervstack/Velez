import {useState} from "react"
import {Link} from "react-router-dom"

import cls from "@/pages/container-registry/components/RegistryInstanceRow/RegistryInstanceRow.module.css"
import {Routes} from "@/app/router/Routes.ts"
import type {RegistryInstance} from "@/app/api/velez"
import StatusDot from "@/components/base/StatusDot.tsx"
import Button from "@/components/base/Button.tsx"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {getLinkToPort} from "@/model/services/VervPlugins.tsx"
import {
    mapRegistryInstanceStatus,
    formatRegistryInstanceCreatedAt,
} from "@/processes/mappings/registry_instances.ts"
import RegistryInstanceCredentials
    from "@/pages/container-registry/components/RegistryInstanceCredentials/RegistryInstanceCredentials.tsx"
import RegistryInstanceDropDialog from "@/dialogs/RegistryInstanceDropDialog/RegistryInstanceDropDialog.tsx"

interface Props {
    instance: RegistryInstance
}

export default function RegistryInstanceRow({instance}: Props) {
    const [expanded, setExpanded] = useState(false)
    const {OpenDialog} = useDialog()

    const name = instance.name ?? ""
    const status = mapRegistryInstanceStatus(instance.status)
    const isRunning = status === "running"
    const hasImageBrowser = !!instance.uiPort && instance.uiPort !== 0 && isRunning

    function handleToggleExpand() {
        setExpanded(!expanded)
    }

    function handleDrop() {
        OpenDialog(<RegistryInstanceDropDialog name={name}/>)
    }

    function handleOpenImageBrowser() {
        window.open(getLinkToPort(instance.uiPort as number), "_blank")
    }

    return (
        <div className={cls.RegistryInstanceRowContainer}>
            <div className={cls.row}>
                <StatusDot status={status} pulse/>
                <Link className={cls.name} to={Routes.Service + "/" + name}>{name}</Link>
                <span className={cls.cell}>{instance.environment || "-"}</span>
                <span className={cls.cell}>{instance.port}</span>
                <span className={cls.cell}>{instance.username}</span>
                <span className={cls.cell}>{formatRegistryInstanceCreatedAt(instance.createdAt)}</span>
                <div className={cls.actions}>
                    {hasImageBrowser && (
                        <Button sm onClick={handleOpenImageBrowser}>
                            Open image browser ↗
                        </Button>
                    )}
                    <Button sm onClick={handleToggleExpand}>
                        {expanded ? "Hide" : "Credentials"}
                    </Button>
                    <Button sm variant="danger" onClick={handleDrop}>
                        Drop
                    </Button>
                </div>
            </div>
            {expanded && (
                <RegistryInstanceCredentials
                    name={name}
                    username={instance.username ?? ""}
                />
            )}
        </div>
    )
}
