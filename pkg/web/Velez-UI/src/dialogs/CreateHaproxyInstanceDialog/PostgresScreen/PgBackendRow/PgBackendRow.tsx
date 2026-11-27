import {Button, Dropdown} from "@vervstack/chures"
import {useState} from "react"

import type {PgBackend} from "@/app/api/velez/haproxy_api.pb"
import {haproxyService, PgBackendRole} from "@/processes/api/haproxy.ts"
import cls from "@/dialogs/CreateHaproxyInstanceDialog/CreateHaproxyInstanceDialog.module.css"

interface Props {
    backend: PgBackend
    sidecarContainerId?: string
    isMaster: boolean
    onRoleChange(role: PgBackendRole): void
    onRemove(): void
}

const ROLE_OPTIONS = [
    {id: PgBackendRole.BACKUP, name: "Backup"},
    {id: PgBackendRole.BALANCED, name: "Balanced"},
]

export default function PgBackendRow({backend, sidecarContainerId, isMaster, onRoleChange, onRemove}: Props) {
    const [isChecking, setIsChecking] = useState(false)
    const [error, setError] = useState("")

    async function retest() {
        setIsChecking(true)
        setError("")

        try {
            const result = await haproxyService.validatePgBackend({backend, sidecarContainerId})
            if (!result.ok) setError(result.errorMessage || "PostgreSQL is unavailable.")
        } catch (failure) {
            setError(failure instanceof Error ? failure.message : String(failure))
        } finally {
            setIsChecking(false)
        }
    }

    return (
        <div className={cls.Backend}>
            <div className={cls.Actions}>
                <strong>{backend.velezService || backend.endpoint?.host}</strong>
                {isMaster && <span className={cls.Master}>Master</span>}
            </div>
            {!isMaster && (
                <Dropdown label="Role" options={ROLE_OPTIONS} value={[backend.role ?? PgBackendRole.BACKUP]}
                          onChange={([id]) => onRoleChange(id === PgBackendRole.BALANCED ?
                              PgBackendRole.BALANCED : PgBackendRole.BACKUP)} portal/>
            )}
            <div className={cls.Actions}>
                <small role="status" className={error ? cls.Unreachable : cls.Reachable}>
                    {isChecking ? "Checking…" : error || "Available"}
                </small>
                <Button variant="secondary" type="button" disabled={isChecking} onClick={retest}>Retest</Button>
                <Button variant="secondary" type="button" onClick={onRemove}>Remove</Button>
            </div>
        </div>
    )
}
