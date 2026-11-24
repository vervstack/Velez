import {useEffect} from "react"

import cls from "@/widgets/service/PgConnectionSection/PgConnectionSection.module.css"

import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import Button from "@/components/base/Button.tsx"
import EnablePgReplicationDialog from "@/dialogs/EnablePgReplicationDialog/EnablePgReplicationDialog.tsx"
import SetupPgReplicationDialog from "@/dialogs/SetupPgReplicationDialog/SetupPgReplicationDialog.tsx"
import {formatPortMapping} from "@/processes/mappings/ports.ts"
import {useListPgInstancesQuery} from "@/processes/queries/pg_instances.ts"
import PgInstanceCredentials from "@/widgets/PgInstanceCredentials/PgInstanceCredentials.tsx"
import PgReplicationStatus from
    "@/widgets/service/PgConnectionSection/components/PgReplicationStatus/PgReplicationStatus.tsx"

interface Props {
    serviceName: string
}

export default function PgConnectionSection({serviceName}: Props) {
    const {OpenDialog} = useDialog()
    const toaster = useToaster()

    const instancesQuery = useListPgInstancesQuery(true)
    useEffect(() => {
        if (instancesQuery.error) toaster.catchGrpc(instancesQuery.error)
    }, [instancesQuery.error])

    const instance = instancesQuery.data?.instances?.find((item) => item.name === serviceName)
    if (!instance) return null

    function handleEnableReplication() {
        OpenDialog(<EnablePgReplicationDialog name={serviceName}/>)
    }

    function handleSetupReplication() {
        OpenDialog(<SetupPgReplicationDialog name={serviceName}/>)
    }

    return (
        <div className={cls.PgConnectionSectionContainer}>
            <span className={cls.SectionTitle}>Connection</span>
            <div className={cls.PortRow}>
                <span className={cls.Label}>Port</span>
                <span className={cls.Value}>{formatPortMapping(instance.port, instance.exposedPort)}</span>
            </div>
            <PgInstanceCredentials
                name={instance.name ?? ""}
                dbName={instance.dbName ?? ""}
                username={instance.username ?? ""}
            />
            <div className={cls.ActionsWrapper}>
                <Button onClick={handleEnableReplication}>Enable replication</Button>
                <Button onClick={handleSetupReplication}>Set up replication</Button>
            </div>
            <PgReplicationStatus instance={serviceName}/>
        </div>
    )
}
