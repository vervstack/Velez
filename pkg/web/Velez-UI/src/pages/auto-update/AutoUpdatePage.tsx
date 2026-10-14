import {useMemo} from "react"

import cls from "@/pages/auto-update/AutoUpdatePage.module.css"
import type {ListAutoUpdateServicesResponseService} from "@/app/api/velez/auto_update_api.pb"
import {describeCheckPeriod, sortServices} from "@/processes/autoUpdate.ts"
import {useListAutoUpdateServicesQuery} from "@/processes/queries/autoUpdate.ts"
import InstanceCount from "@/components/InstanceCount/InstanceCount.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import AutoUpdateDisabledBanner
    from "@/pages/auto-update/components/AutoUpdateDisabledBanner/AutoUpdateDisabledBanner.tsx"
import AutoUpdateRow from "@/pages/auto-update/components/AutoUpdateRow/AutoUpdateRow.tsx"
import AutoUpdateTableSkeleton
    from "@/pages/auto-update/components/AutoUpdateTableSkeleton/AutoUpdateTableSkeleton.tsx"

const COLUMNS = ["Name", "Image", "Bump level", "Enabled"]

function renderColumnHeader(label: string) {
    return <span key={label} className={cls.HeaderCell}>{label}</span>
}

function renderRow(service: ListAutoUpdateServicesResponseService) {
    return <AutoUpdateRow key={service.name} service={service}/>
}

export default function AutoUpdatePage() {
    const servicesQuery = useListAutoUpdateServicesQuery()

    const services = useMemo(() => sortServices(servicesQuery.data?.services ?? []), [servicesQuery.data])
    const watchedCount = useMemo(() => services.filter((s) => s.isEnabled).length, [services])
    const isWorkerOff = servicesQuery.isSuccess && servicesQuery.data?.isWorkerEnabled === false

    function handleRetry() {
        servicesQuery.refetch()
    }

    function renderContent() {
        if (servicesQuery.isLoading) return <AutoUpdateTableSkeleton/>
        if (servicesQuery.isError) {
            return <QueryErrorState message="Failed to load auto-update services." onRetry={handleRetry}/>
        }
        if (services.length === 0) {
            return <div className={cls.EmptyMessage}>No Velez app services on this node.</div>
        }
        return (
            <div className={cls.Table}>
                <div className={cls.TableHeader}>
                    {COLUMNS.map(renderColumnHeader)}
                </div>
                {services.map(renderRow)}
            </div>
        )
    }

    return (
        <div className={cls.AutoUpdatePageContainer}>
            <div className={cls.Toolbar}>
                <h1 className={cls.PageTitle}>Auto-update</h1>
                <InstanceCount count={watchedCount} label="watched" isLoading={servicesQuery.isLoading}/>
                <span className={cls.Period}>
                    {servicesQuery.isSuccess && describeCheckPeriod(servicesQuery.data?.checkPeriodSeconds)}
                </span>
            </div>
            {isWorkerOff && <AutoUpdateDisabledBanner/>}
            {renderContent()}
        </div>
    )
}
