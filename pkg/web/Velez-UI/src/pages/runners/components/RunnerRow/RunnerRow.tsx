import {useState} from "react"

import cls from "@/pages/runners/components/RunnerRow/RunnerRow.module.css"
import type {Runner} from "@/app/api/velez"
import StatusDot from "@/components/base/StatusDot.tsx"
import Button from "@/components/base/Button.tsx"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import RunnerDropDialog from "@/dialogs/RunnerDropDialog/RunnerDropDialog.tsx"
import RunnerCredentials from "@/pages/runners/components/RunnerCredentials/RunnerCredentials.tsx"

interface Props {
    runner: Runner
}

export default function RunnerRow({runner}: Props) {
    const [expanded, setExpanded] = useState(false)
    const {OpenDialog} = useDialog()

    const name = runner.name ?? ""

    function handleToggleExpand() {
        setExpanded(!expanded)
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
                <span className={cls.name}>{name}</span>
                <span className={cls.cell}>{runner.provider || "-"}</span>
                <span className={cls.cell}>{runner.scope || "-"}</span>
                <span className={cls.cell}>{runner.target || "-"}</span>
                <span className={cls.cell}>{runner.status || "-"}</span>
                <span className={cls.cell}>
                    {runner.createdAt ? new Date(runner.createdAt as never).toLocaleDateString() : "-"}
                </span>
                <div className={cls.actions}>
                    <Button sm onClick={handleToggleExpand}>
                        {expanded ? "Hide" : "Credentials"}
                    </Button>
                    <Button sm variant="danger" onClick={handleDrop}>
                        Drop
                    </Button>
                </div>
            </div>
            {expanded && <RunnerCredentials name={name}/>}
        </div>
    )
}
