import cls from "@/widgets/settings/SandboxSettings/SandboxSettings.module.css"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import type {UpdateSettingsRequest} from "@/app/api/velez/settings_api.pb"
import {UpdateSettingsMutation, useGetSettingsQuery} from "@/processes/queries/settings.ts"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import SysboxSetupChecklist from "@/widgets/settings/SandboxSettings/components/SysboxSetupChecklist/SysboxSetupChecklist.tsx"
import SandboxToggleRow from "@/widgets/settings/SandboxSettings/components/SandboxToggleRow/SandboxToggleRow.tsx"

const SYSBOX_DESCRIPTION =
    "All containers started by Velez use the sysbox-runc runtime unless whitelisted for elevated access. " +
    "Requires Sysbox installed on the host."

const WHITELIST_IGNORED_DESCRIPTION =
    "Sysbox ignores the whitelist, privileged system containers such as the VPN sidecar may fail to start."

export default function SandboxSettings() {
    const toaster = useToaster()
    const settingsQuery = useGetSettingsQuery()
    const updateSettings = UpdateSettingsMutation()

    const settings = settingsQuery.data?.settings

    function handleUpdate(req: UpdateSettingsRequest) {
        updateSettings.mutate(req, {onError: toaster.catchGrpc})
    }

    function handleSysboxChange(isChecked: boolean) {
        handleUpdate({isSysboxEnabled: isChecked})
    }

    function handleWhitelistIgnoredChange(isChecked: boolean) {
        handleUpdate({isSysboxWhitelistIgnored: isChecked})
    }

    function handleRetry() {
        settingsQuery.refetch()
    }

    function renderContent() {
        if (settingsQuery.isLoading) {
            return <SkeletonLoader shape="block" height="5rem"/>
        }
        if (settingsQuery.isError) {
            return <QueryErrorState message="Failed to load settings." onRetry={handleRetry}/>
        }
        return (
            <>
                <SandboxToggleRow
                    label="Run containers in Sysbox"
                    description={SYSBOX_DESCRIPTION}
                    isChecked={settings?.isSysboxEnabled ?? false}
                    isDisabled={updateSettings.isPending}
                    onChange={handleSysboxChange}
                />
                <SandboxToggleRow
                    label="Apply Sysbox to whitelisted containers too"
                    description={WHITELIST_IGNORED_DESCRIPTION}
                    isChecked={settings?.isSysboxWhitelistIgnored ?? false}
                    isDisabled={updateSettings.isPending}
                    onChange={handleWhitelistIgnoredChange}
                />
            </>
        )
    }

    return (
        <div className={cls.SandboxSettingsContainer}>
            <div className={cls.SectionTitle}>Sandbox</div>
            {renderContent()}
            <SysboxSetupChecklist/>
        </div>
    )
}
