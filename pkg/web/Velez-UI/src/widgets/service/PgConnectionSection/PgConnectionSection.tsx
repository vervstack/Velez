import {useEffect} from "react"

import cls from "@/widgets/service/PgConnectionSection/PgConnectionSection.module.css"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {formatPortMapping} from "@/processes/mappings/ports.ts"
import {useListPgInstancesQuery} from "@/processes/queries/pg_instances.ts"
import PgInstanceCredentials from "@/widgets/PgInstanceCredentials/PgInstanceCredentials.tsx"

interface Props {
    serviceName: string
}

export default function PgConnectionSection({serviceName}: Props) {
    const toaster = useToaster()

    const instancesQuery = useListPgInstancesQuery()
    useEffect(() => {
        if (instancesQuery.error) toaster.catchGrpc(instancesQuery.error)
    }, [instancesQuery.error])

    const instance = instancesQuery.data?.instances?.find((item) => item.name === serviceName)
    if (!instance) return null

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
        </div>
    )
}
