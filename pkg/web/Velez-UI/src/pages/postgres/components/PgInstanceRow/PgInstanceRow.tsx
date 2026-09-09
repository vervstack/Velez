import {useState} from "react"
import {Link} from "react-router-dom"

import cls from "@/pages/postgres/components/PgInstanceRow/PgInstanceRow.module.css"
import {Routes} from "@/app/router/Routes.ts"
import type {PgInstance} from "@/app/api/velez"
import StatusDot from "@/components/base/StatusDot.tsx"
import Button from "@/components/base/Button.tsx"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {mapPgInstanceStatus, formatPgInstanceCreatedAt} from "@/processes/mappings/pg_instances.ts"
import PgInstanceCredentials from "@/pages/postgres/components/PgInstanceCredentials/PgInstanceCredentials.tsx"
import PgInstanceDropDialog from "@/dialogs/PgInstanceDropDialog/PgInstanceDropDialog.tsx"

interface Props {
    instance: PgInstance
}

export default function PgInstanceRow({instance}: Props) {
    const [expanded, setExpanded] = useState(false)
    const {OpenDialog} = useDialog()

    const name = instance.name ?? ""

    function handleToggleExpand() {
        setExpanded(!expanded)
    }

    function handleDrop() {
        OpenDialog(<PgInstanceDropDialog name={name}/>)
    }

    return (
        <div className={cls.PgInstanceRowContainer}>
            <div className={cls.row}>
                <StatusDot status={mapPgInstanceStatus(instance.status)} pulse/>
                <Link className={cls.name} to={Routes.Service + "/" + name}>{name}</Link>
                <span className={cls.cell}>{instance.environment || "-"}</span>
                <span className={cls.cell}>{instance.dbName}</span>
                <span className={cls.cell}>{instance.username}</span>
                <span className={cls.cell}>{instance.port}</span>
                <span className={cls.cell}>{formatPgInstanceCreatedAt(instance.createdAt)}</span>
                <div className={cls.actions}>
                    <Button sm onClick={handleToggleExpand}>
                        {expanded ? "Hide" : "Credentials"}
                    </Button>
                    <Button sm variant="danger" onClick={handleDrop}>
                        Drop
                    </Button>
                </div>
            </div>
            {expanded && (
                <PgInstanceCredentials
                    name={name}
                    dbName={instance.dbName ?? ""}
                    username={instance.username ?? ""}
                />
            )}
        </div>
    )
}
