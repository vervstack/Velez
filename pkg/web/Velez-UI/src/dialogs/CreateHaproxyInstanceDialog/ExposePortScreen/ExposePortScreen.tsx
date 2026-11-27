import {useEffect, useMemo} from "react"
import {Checkbox, Dropdown} from "@vervstack/chures"

import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import type {StepScreenProps} from "@/dialogs/StepsDialog/Step.ts"
import {expandAvailablePorts} from "@/processes/availablePorts.ts"
import {useVersionQuery} from "@/processes/queries/version.ts"
import {
    DEFAULT_PORT,
    type HaproxyWizardContext,
    parsePort,
} from "@/dialogs/CreateHaproxyInstanceDialog/haproxyWizardContext.ts"
import cls from "@/dialogs/CreateHaproxyInstanceDialog/CreateHaproxyInstanceDialog.module.css"

type ScreenProps = StepScreenProps<HaproxyWizardContext>

export default function ExposePortScreen(props: ScreenProps) {
    const {isPortExposed, port, sidecarContainerId, isPortReady: wasPortReady, updateContext} = props
    const version = useVersionQuery()
    const ranges = version.data?.availablePorts
    const ports = useMemo(() => expandAvailablePorts(ranges ?? []), [ranges])
    const options = useMemo(() => Array.from(ports, (number) => ({
        id: String(number), name: String(number),
    })), [ports])
    const firstPort = ports.values().next().value
    const suggestedPort = ports.has(DEFAULT_PORT) ? DEFAULT_PORT : firstPort
    const portNumber = parsePort(port ?? "")
    const isPortReady = portNumber !== undefined && ports.has(portNumber)

    useEffect(() => {
        if (sidecarContainerId) return

        if (isPortExposed && port === undefined && suggestedPort !== undefined) {
            updateContext({port: String(suggestedPort), isPortReady: true})
        } else if (isPortReady !== wasPortReady) {
            updateContext({isPortReady})
        }
    }, [isPortExposed, isPortReady, port, sidecarContainerId, suggestedPort, updateContext, wasPortReady])

    function handlePortChange(ids: string[]) {
        const value = ids[0] ?? ""
        const number = parsePort(value)
        updateContext({port: value, isPortReady: number !== undefined && ports.has(number)})
    }

    return (
        <div className={cls.Fields}>
            <Checkbox label="Expose port" checked={isPortExposed} disabled={Boolean(sidecarContainerId)}
                      onChange={(isChecked) => updateContext({isPortExposed: isChecked, isPortReady})}/>
            {isPortExposed && sidecarContainerId && <span>Published port {port}</span>}
            {isPortExposed && !sidecarContainerId && version.isLoading && <SkeletonLoader shape="line"/>}
            {isPortExposed && !sidecarContainerId && version.isError && (
                <QueryErrorState message="Failed to load the node's port pool." onRetry={version.refetch}/>
            )}
            {isPortExposed && !sidecarContainerId && version.isSuccess && (
                <>
                    <Dropdown label="Published port" placeholder="Select a port" options={options}
                              value={port ? [port] : []} onChange={handlePortChange} portal/>
                    {ports.size === 0 && <small className={cls.Unreachable}>The node has no configured ports.</small>}
                </>
            )}
        </div>
    )
}
