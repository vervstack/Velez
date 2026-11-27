import {useEffect} from "react"
import {Button, Input} from "@vervstack/chures"

import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import {
    type VcnNetworkWizardContext,
    validateCustomPort,
} from "@/dialogs/CreateVcnNetworkDialog/processes/vcnNetworkWizard.ts"
import type {StepScreenProps} from "@/dialogs/StepsDialog/Step.ts"
import {useNetworkStatusQuery} from "@/processes/queries/networks.ts"
import cls from "@/dialogs/CreateVcnNetworkDialog/screens/NetworkSourceScreen/NetworkSourceScreen.module.css"

type Props = StepScreenProps<VcnNetworkWizardContext>

export default function NetworkSourceScreen({source, customPort, customImage, url, token, updateContext}: Props) {
    const statusQuery = useNetworkStatusQuery()
    const canConnectExternal = statusQuery.data?.isClusterMode === true
    const externalTooltip = statusQuery.isSuccess && !canConnectExternal
        ? "External networks require cluster mode"
        : undefined

    useEffect(() => {
        updateContext({isExternalAvailable: canConnectExternal})
    }, [canConnectExternal, updateContext])

    function chooseDeploy() {
        updateContext({source: "deploy"})
    }

    function chooseExternal() {
        if (canConnectExternal) updateContext({source: "external"})
    }

    return (
        <div className={cls.NetworkSourceScreenContainer}>
            <div className={cls.OptionsWrapper} role="group" aria-label="Network source">
                <Button type="button" variant={source === "deploy" ? "primary" : "secondary"}
                        aria-pressed={source === "deploy"} onClick={chooseDeploy}>
                    Deploy new Headscale
                </Button>
                <Button type="button" variant={source === "external" ? "primary" : "secondary"}
                        className={cls.ExternalOption} aria-pressed={source === "external"}
                        aria-disabled={!canConnectExternal} aria-description={externalTooltip}
                        data-tooltip-id={externalTooltip ? "root-tooltip" : undefined}
                        data-tooltip-content={externalTooltip} onClick={chooseExternal}>
                    Connect to external
                </Button>
            </div>
            {statusQuery.isLoading && <SkeletonLoader shape="line"/>}
            {statusQuery.isError && (
                <QueryErrorState message="Could not verify cluster mode." onRetry={statusQuery.refetch}/>
            )}
            {source === "deploy" ? (
                <>
                    <Input label="Custom port (optional)" inputMode="numeric" value={customPort}
                           setValue={(value) => updateContext({customPort: value})}
                           error={validateCustomPort(customPort)} placeholder="Automatically assigned"/>
                    <Input label="Custom image (optional)" value={customImage}
                           setValue={(value) => updateContext({customImage: value})} placeholder="Default image"/>
                </>
            ) : (
                <>
                    <Input label="Headscale URL" value={url} placeholder="https://headscale.example.com"
                           setValue={(value) => updateContext({url: value})} disabled={!canConnectExternal}/>
                    <Input label="API token" type="password" value={token} autoComplete="off"
                           setValue={(value) => updateContext({token: value})} disabled={!canConnectExternal}/>
                </>
            )}
        </div>
    )
}
