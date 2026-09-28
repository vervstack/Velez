import {useNavigate} from "react-router-dom"

import cls from "@/pages/runners/components/RunnerRow/RunnerRow.module.css"
import type {Runner} from "@/app/api/velez"
import StatusDot from "@/components/base/StatusDot.tsx"
import Button from "@/components/base/Button.tsx"
import GitlabIcon from "@/components/base/icons/GitlabIcon.tsx"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {deriveRunnerDisplayName, isGitlabRunnerName} from "@/processes/mappings/runnerDisplay.ts"
import RunnerDropDialog from "@/dialogs/RunnerDropDialog/RunnerDropDialog.tsx"
import {Routes} from "@/app/router/Routes.ts"

interface Props {
    runner: Runner
}

export default function RunnerRow({runner}: Props) {
    const {OpenDialog} = useDialog()
    const navigate = useNavigate()

    const name = runner.name ?? ""

    function handleOpenService() {
        navigate(Routes.Service + "/" + name)
    }

    function handleDrop() {
        OpenDialog(<RunnerDropDialog name={name}/>)
    }

    function mapRunnerStatus(status?: string): "running" | "stopped" | "pending" | "error" {
        if (status === "running") return "running"
        if (status === "stopped") return "stopped"
        if (status === "error") return "error"
        return "pending"
    }

    return (
        <div className={cls.RunnerRowContainer}>
            <div className={cls.row}>
                <StatusDot status={mapRunnerStatus(runner.status)} pulse/>
                {isGitlabRunnerName(name) && <GitlabIcon className={cls.icon}/>}
                <span className={cls.name} onClick={handleOpenService}>{deriveRunnerDisplayName(name)}</span>
                <span className={cls.cell}>{runner.provider || "-"}</span>
                <span className={cls.cell}>{runner.scope || "-"}</span>
                <span className={cls.cell}>{runner.target || "-"}</span>
                <span className={cls.cell}>{runner.status || "-"}</span>
                <span className={cls.cell}>
                    {runner.createdAt ? new Date(runner.createdAt as never).toLocaleDateString() : "-"}
                </span>
                <div className={cls.actions}>
                    <Button sm variant="danger" onClick={handleDrop}>
                        Drop
                    </Button>
                </div>
            </div>
        </div>
    )
}
