import {useServiceUpgrade} from "@/app/hooks/serviceUpgrade/ServiceUpgrade.ts"
import {UpgradeProgress, upgradeProgress} from "@/processes/upgradeProgress.ts"

export function useUpgradeProgress(serviceName: string): UpgradeProgress {
    const status = useServiceUpgrade((state) => state.statusByService[serviceName])

    return upgradeProgress(status)
}
