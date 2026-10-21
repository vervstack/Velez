import {useMemo} from "react"

import cls from "@/dialogs/CreateServiceDialog/components/NetworkGroupList/NetworkGroupList.module.css"
import type {DockerContainer} from "@/app/api/velez"
import {useEnvironmentStore} from "@/app/hooks/environment/Environment.ts"
import {useRegisterPreviewQuery} from "@/processes/queries/containers.ts"
import NetworkGroupItem from "@/dialogs/CreateServiceDialog/components/NetworkGroupItem/NetworkGroupItem.tsx"
import type {AdoptGroup} from "@/dialogs/CreateServiceDialog/hooks/AdoptGroup.ts"
import {RegisterPattern} from "@/dialogs/CreateServiceDialog/processes/buildRegisterContainerRequest.ts"
import {buildPreviewRequest} from "@/dialogs/CreateServiceDialog/processes/buildPreviewRequest.ts"
import {previewOfContainer} from "@/dialogs/CreateServiceDialog/processes/registerPreview.ts"
import {isGenericOnly, serviceIdOf} from "@/dialogs/CreateServiceDialog/processes/registerRoles.ts"
import {ServiceScreen, serviceTypeLabel} from "@/dialogs/CreateServiceDialog/processes/serviceScreen.ts"

interface Props {
    group: AdoptGroup
    pattern: RegisterPattern
    serviceScreen?: ServiceScreen
    onPickService(containerId: string): void
    onRequestRoot(containerId: string): void
}

const GENERIC_ONLY_NOTE = "The service or network root differs from the original container, so only the Generic " +
    "pattern is available and ports, bind-mount links and the image tag are left as they are."

export default function NetworkGroupList({group, pattern, serviceScreen, onPickService, onRequestRoot}: Props) {
    const environment = useEnvironmentStore((state) => state.selectedEnvironment)
    const {containers, originalRootId, roles, serviceName} = group

    const request = useMemo(() => buildPreviewRequest({
        containerId: originalRootId,
        environment,
        serviceName,
        pattern,
        roles,
    }), [originalRootId, environment, serviceName, pattern, roles])
    const query = useRegisterPreviewQuery(request)

    const serviceId = serviceIdOf(roles)
    const isGeneric = isGenericOnly(roles, originalRootId)
    const typeLabel = serviceTypeLabel(isGeneric ? "generic" : serviceScreen)
    const serviceLabel = serviceName.trim() || "the service"

    function renderItem(container: DockerContainer) {
        const containerId = container.id ?? ""
        const isService = containerId === serviceId

        const preview = {
            preview: previewOfContainer(query.data, containerId),
            isLoading: query.isLoading,
            isError: query.isError,
            isDisabled: request === null,
            onRetry: query.refetch,
        }

        return (
            <NetworkGroupItem
                key={containerId}
                container={container}
                roles={{isRoot: containerId === roles.rootId, isService, canSwitch: containers.length > 1}}
                description={isService ? typeLabel : `Resource of ${serviceLabel}`}
                preview={preview}
                onMakeService={onPickService}
                onMakeRoot={onRequestRoot}
            />
        )
    }

    return (
        <div className={cls.NetworkGroupListContainer}>
            <span className={cls.Legend}>
                <span className={cls.Swatch}/>
                <span className={cls.LegendLabel}>Transferred from the container</span>
                <span className={cls.AddedSwatch}/>
                <span className={cls.LegendLabel}>Added by onboarding</span>
            </span>
            {isGeneric && <span className={cls.Note}>{GENERIC_ONLY_NOTE}</span>}
            {request === null && <span className={cls.Note}>Enter a service name to preview the result.</span>}
            <div className={cls.ItemsWrapper}>
                {containers.map(renderItem)}
            </div>
        </div>
    )
}
