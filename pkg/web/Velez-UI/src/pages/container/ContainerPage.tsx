import {useEffect} from "react"
import {useNavigate, useParams, useSearchParams} from "react-router-dom"

import cls from "@/pages/container/ContainerPage.module.css"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import {useBreadcrumbs} from "@/app/hooks/breadcrumbs/Breadcrumbs.ts"
import {useGetContainerQuery} from "@/processes/queries/containers.ts"
import {deriveContainerServicePresentation} from "@/processes/mappings/containers.ts"
import {CONTAINER_TABS, ContainerTab, parseContainerTab} from "@/processes/mappings/containerTabs.ts"
import Button from "@/components/base/Button.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import ServiceChip from "@/components/base/chips/ServiceChip.tsx"
import {Routes} from "@/app/router/Routes.ts"
import ContainerDetailsTab from "@/pages/container/components/ContainerDetailsTab/ContainerDetailsTab.tsx"
import ContainerInspectTab from "@/pages/container/components/ContainerInspectTab/ContainerInspectTab.tsx"
import ContainerLogsTab from "@/pages/container/components/ContainerLogsTab/ContainerLogsTab.tsx"
import ContainerPageSkeleton from "@/pages/container/components/ContainerPageSkeleton/ContainerPageSkeleton.tsx"
import ContainerStatusBadge from "@/pages/container/components/ContainerStatusBadge/ContainerStatusBadge.tsx"

export default function ContainerPage() {
    const params = useParams<Record<string, string>>()
    const navigate = useNavigate()
    const toaster = useToaster()
    const [searchParams, setSearchParams] = useSearchParams()
    const containerId = params["id"] || ""
    const activeTab = parseContainerTab(searchParams.get("tab"))

    const {data: container, isLoading, isError, error, refetch} = useGetContainerQuery(
        containerId, containerId !== ""
    )
    useEffect(() => {
        if (error) toaster.catchGrpc(error)
    }, [error])

    const setCrumbs = useBreadcrumbs((s) => s.setCrumbs)
    const linkedServiceName = container?.linkedServiceName

    useEffect(function publishBreadcrumbs() {
        if (containerId === "") return
        const containerCrumb = {label: container?.name || containerId}

        if (linkedServiceName) {
            setCrumbs([
                {label: "services", onClick: () => navigate("/")},
                {label: linkedServiceName, onClick: () => navigate(Routes.Service + "/" + linkedServiceName)},
                containerCrumb,
            ])
        } else {
            setCrumbs([{label: "Home", onClick: () => navigate("/")}, containerCrumb])
        }
        return function clearBreadcrumbs() {
            setCrumbs([])
        }
    }, [containerId, container?.name, linkedServiceName, setCrumbs, navigate])

    function renderTabButton(item: {id: ContainerTab, label: string}) {
        function handleClick() {
            setSearchParams({tab: item.id}, {replace: true})
        }

        return (
            <Button key={item.id} sm variant={activeTab === item.id ? "primary" : "secondary"} onClick={handleClick}>
                {item.label}
            </Button>
        )
    }

    if (!containerId) {
        return <div className={cls.ContainerPageContainer}>
            <div className={cls.ErrorMessage}>No id provided.</div>
        </div>
    }

    if (isLoading) {
        return <ContainerPageSkeleton/>
    }

    if (isError) {
        return (
            <div className={cls.ContainerPageContainer}>
                <QueryErrorState message="Failed to load container." onRetry={refetch}/>
            </div>
        )
    }

    if (!container) {
        return <div className={cls.ContainerPageContainer}>
            <div className={cls.ErrorMessage}>Container not found.</div>
        </div>
    }

    const servicePresentation = container.linkedServiceName
        ? deriveContainerServicePresentation(container)
        : null

    return (
        <div className={cls.ContainerPageContainer}>
            <div className={cls.Header}>
                <div className={cls.ContainerNameWrapper}>
                    <div className={cls.ContainerName}>{container.name || container.id}</div>
                    <ContainerStatusBadge status={container.status}/>
                    {!container.isRegistered && <ContainerStatusBadge isUnregistered/>}
                </div>
                {servicePresentation && (
                    <ServiceChip
                        serviceName={servicePresentation.displayName}
                        icon={servicePresentation.icon}
                        onClick={() => navigate(Routes.Service + "/" + container.linkedServiceName)}
                    />
                )}
            </div>

            <div className={cls.Tabs}>
                {CONTAINER_TABS.map(renderTabButton)}
            </div>

            {activeTab === "details" && <ContainerDetailsTab container={container}/>}
            {activeTab === "logs" && <ContainerLogsTab containerId={containerId}/>}
            {activeTab === "inspect" && <ContainerInspectTab container={container}/>}
        </div>
    )
}
