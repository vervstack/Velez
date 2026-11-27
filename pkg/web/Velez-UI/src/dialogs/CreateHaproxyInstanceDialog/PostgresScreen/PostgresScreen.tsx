import {useEffect} from "react"
import {Dropdown} from "@vervstack/chures"

import Input from "@/components/base/Input.tsx"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import TargetStatus from "@/dialogs/CreateHaproxyInstanceDialog/TargetStatus/TargetStatus.tsx"
import type {StepScreenProps} from "@/dialogs/StepsDialog/Step.ts"
import {validateInstanceName} from "@/processes/mappings/instanceName.ts"
import {useDockerHaproxyTargetQuery} from "@/processes/queries/haproxy.ts"
import {useListPgInstancesQuery} from "@/processes/queries/pg_instances.ts"
import {DEFAULT_PORT, type HaproxyWizardContext} from "@/dialogs/CreateHaproxyInstanceDialog/haproxyWizardContext.ts"
import cls from "@/dialogs/CreateHaproxyInstanceDialog/CreateHaproxyInstanceDialog.module.css"

type ScreenProps = StepScreenProps<HaproxyWizardContext>

export default function PostgresScreen({name, master, environment, box, dockerNetwork, updateContext}: ScreenProps) {
    const pgQuery = useListPgInstancesQuery(true)
    const target = useDockerHaproxyTargetQuery(master, environment)
    const pgInstances = pgQuery.data?.instances ?? []
    const options = pgInstances.filter((pg) => pg.name).map((pg) => ({
        id: pg.name ?? "",
        name: pg.environment ? `${pg.displayName || pg.name} (${pg.environment})` : pg.displayName || pg.name || "",
    }))

    useEffect(() => {
        const detectedNetwork = target.data?.dockerNetwork ?? ""
        if (dockerNetwork !== detectedNetwork) updateContext({dockerNetwork: detectedNetwork})
    }, [dockerNetwork, target.data?.dockerNetwork, updateContext])

    function handleMasterChange(ids: string[]) {
        const pg = pgInstances.find((instance) => instance.name === ids[0])
        updateContext({
            master: pg?.name ?? "",
            masterLabel: options.find((option) => option.id === pg?.name)?.name ?? "",
            environment: pg?.environment ?? "",
            dockerNetwork: "",
            mode: undefined,
            host: "",
            tailscaleIp: "",
            targetPort: String(pg?.exposedPort || DEFAULT_PORT),
            isNetworkReady: false,
        })
    }

    return (
        <div className={cls.Fields}>
            <Input label="HAProxy name" inputValue={name} onChange={(value) => updateContext({name: value})}
                   error={validateInstanceName(name.trim())}/>
            {pgQuery.isLoading && <SkeletonLoader shape="line"/>}
            {pgQuery.isError && (
                <QueryErrorState message="Failed to load PostgreSQL instances." onRetry={pgQuery.refetch}/>
            )}
            {pgQuery.isSuccess && (
                <Dropdown label="PostgreSQL instance" placeholder="Select a PostgreSQL instance" options={options}
                          value={master ? [master] : []} onChange={handleMasterChange} portal/>
            )}
            {master && <TargetStatus query={target}/>}
            <Input label="Box (optional)" inputValue={box} onChange={(value) => updateContext({box: value})}/>
        </div>
    )
}
