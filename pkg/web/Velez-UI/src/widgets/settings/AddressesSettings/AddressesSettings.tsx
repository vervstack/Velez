import {useEffect, useRef} from "react"

import cls from "@/widgets/settings/AddressesSettings/AddressesSettings.module.css"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {RebuildAddressesMutation, useAddressesRebuildStatusQuery} from "@/processes/queries/settings.ts"
import Button from "@/components/base/Button.tsx"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import RebuildProgress from "@/widgets/settings/AddressesSettings/components/RebuildProgress/RebuildProgress.tsx"

export default function AddressesSettings() {
    const toaster = useToaster()
    const statusQuery = useAddressesRebuildStatusQuery()
    const rebuildAddresses = RebuildAddressesMutation()

    const status = statusQuery.data
    const isRunning = status?.isRunning ?? false
    const lastError = status?.lastError ?? ""
    const wasRunning = useRef(false)

    useEffect(() => {
        if (wasRunning.current && !isRunning && lastError === "") {
            toaster.bake({title: "Addresses rebuilt", description: "", level: "Info"})
        }
        wasRunning.current = isRunning
    }, [isRunning, lastError, toaster])

    function handleRebuild() {
        rebuildAddresses.mutate(undefined, {onError: toaster.catchGrpc})
    }

    function handleRetry() {
        statusQuery.refetch()
    }

    function renderProgress() {
        if (statusQuery.isLoading) {
            return <SkeletonLoader shape="line" height="1.5rem"/>
        }
        if (statusQuery.isError || !status) {
            return <QueryErrorState message="Failed to load the rebuild status." onRetry={handleRetry}/>
        }
        return <RebuildProgress status={status}/>
    }

    return (
        <div className={cls.AddressesSettingsContainer}>
            <div className={cls.SectionTitle}>Addresses</div>
            <div className={cls.Description}>
                Re-crawl Docker and the VCN to refresh the links shown for service resources.
            </div>
            <div className={cls.ActionsRow}>
                <Button onClick={handleRebuild} disabled={isRunning || rebuildAddresses.isPending}>
                    Rebuild addresses
                </Button>
            </div>
            {renderProgress()}
        </div>
    )
}
