import {useEffect, useMemo, useState} from "react"

import cls from "@/dialogs/CreateServiceDialog/components/AdoptForm/AdoptForm.module.css"
import type {DockerContainer} from "@/app/api/velez"
import {useEnvironmentStore} from "@/app/hooks/environment/Environment.ts"
import {useAdoptGroup} from "@/dialogs/CreateServiceDialog/hooks/AdoptGroup.ts"
import {IsStatefullModeEnabled} from "@/processes/queries/control_plane.ts"
import Input from "@/components/base/Input.tsx"
import BindMountLinks from "@/dialogs/CreateServiceDialog/components/BindMountLinks/BindMountLinks.tsx"
import ImageVersionPicker from "@/dialogs/CreateServiceDialog/components/ImageVersionPicker/ImageVersionPicker.tsx"
import NoPortsNotice from "@/dialogs/CreateServiceDialog/components/NoPortsNotice/NoPortsNotice.tsx"
import PgLoginFields from "@/dialogs/CreateServiceDialog/components/PgLoginFields/PgLoginFields.tsx"
import PortMappingTable from "@/dialogs/CreateServiceDialog/components/PortMappingTable/PortMappingTable.tsx"
import RegistryLoginFields from "@/dialogs/CreateServiceDialog/components/RegistryLoginFields/RegistryLoginFields.tsx"
import RestartNotice from "@/dialogs/CreateServiceDialog/components/RestartNotice/RestartNotice.tsx"
import RunnerFields from "@/dialogs/CreateServiceDialog/components/RunnerFields/RunnerFields.tsx"
import type {OnboardFormState} from "@/dialogs/CreateServiceDialog/processes/onboardContext.ts"
import {isGenericOnly, roleIdsOf, serviceIdOf} from "@/dialogs/CreateServiceDialog/processes/registerRoles.ts"
import {currentTagOf, imageTagToSend} from "@/dialogs/CreateServiceDialog/processes/imageVersion.ts"
import {bindMountsOf, resolveLinks} from "@/dialogs/CreateServiceDialog/processes/bindMounts.ts"
import {
    buildRegisterContainerRequest,
    RegisterPattern,
    RunnerForm,
} from "@/dialogs/CreateServiceDialog/processes/buildRegisterContainerRequest.ts"
import {
    hasVolumes,
    parsePortMappingRows,
    PortMappingRow,
    portMappingRowsOf,
} from "@/dialogs/CreateServiceDialog/processes/portMapping.ts"

interface Props {
    container: DockerContainer
    pattern: RegisterPattern
    isPgLoginRequired?: boolean
    isRegistryLoginRequired?: boolean
    initialRunner?: RunnerForm
    isRegistrationTokenFound?: boolean
    onFormChange(form: OnboardFormState): void
}

export default function AdoptForm({
    container,
    pattern,
    isPgLoginRequired = false,
    isRegistryLoginRequired = false,
    initialRunner,
    isRegistrationTokenFound = false,
    onFormChange,
}: Props) {
    const group = useAdoptGroup()
    const [localServiceName, setLocalServiceName] = useState(container.name ?? "")
    const [linkOverrides, setLinkOverrides] = useState<Record<string, string>>({})
    const [portRows, setPortRows] = useState<PortMappingRow[]>(() => portMappingRowsOf(container))
    const [superuser, setSuperuser] = useState("")
    const [password, setPassword] = useState("")
    const [registryUsername, setRegistryUsername] = useState("")
    const [registryPassword, setRegistryPassword] = useState("")
    const [runner, setRunner] = useState<RunnerForm | undefined>(initialRunner)
    const [selectedTag, setSelectedTag] = useState<string | undefined>(undefined)

    const environment = useEnvironmentStore((state) => state.selectedEnvironment)
    const isClusterMode = IsStatefullModeEnabled()

    const serviceName = group ? group.serviceName : localServiceName
    const setServiceName = group ? group.onServiceNameChange : setLocalServiceName
    const originalRootId = container.id ?? ""
    const isRolesChanged = group ? isGenericOnly(group.roles, originalRootId) : false
    const roleIds = group ? roleIdsOf(group.roles, originalRootId) : {}
    const isRecreating = !isClusterMode || Boolean(roleIds.rootContainerId)
    const serviceContainer = group?.containers.find((member) => member.id === serviceIdOf(group.roles)) ?? container
    const isLinkingShown = !isRolesChanged || isClusterMode

    const links = useMemo(
        () => resolveLinks(bindMountsOf(serviceContainer), serviceName, linkOverrides),
        [serviceContainer, serviceName, linkOverrides],
    )
    const parsedPorts = useMemo(() => parsePortMappingRows(portRows), [portRows])

    const hasPublishedPorts = portRows.length > 0
    const isPortsEditable = !isClusterMode && hasPublishedPorts && !isRolesChanged
    const isRowsInvalid = isPortsEditable && parsedPorts === null

    const request = useMemo(() => {
        if (isRowsInvalid) return null

        return buildRegisterContainerRequest({
            containerId: container.id ?? "",
            environment,
            serviceName,
            pattern,
            pgLogin: isPgLoginRequired ? {superuser, password} : undefined,
            registryLogin: isRegistryLoginRequired
                ? {username: registryUsername, password: registryPassword}
                : undefined,
            runner,
            links,
            isClusterMode,
            isKeepingPorts: false,
            ports: isPortsEditable ? (parsedPorts ?? []) : [],
            imageTag: selectedTag && imageTagToSend(selectedTag, currentTagOf(container.imageName ?? "")),
            ...roleIds,
        })
    }, [container.id, environment, serviceName, pattern, isPgLoginRequired, superuser, password,
        roleIds.rootContainerId, roleIds.serviceContainerId, isRegistryLoginRequired, registryUsername,
        registryPassword, runner, links, isClusterMode, isPortsEditable, parsedPorts, isRowsInvalid, selectedTag,
        container.imageName])

    function handleVolumeNameChange(destination: string, volumeName: string) {
        setLinkOverrides({...linkOverrides, [destination]: volumeName})
    }

    useEffect(() => {
        onFormChange({request, isRecreating, hasPublishedPorts})
    }, [request, isRecreating, hasPublishedPorts])

    return (
        <div className={cls.AdoptFormContainer}>
            <div className={cls.FieldsWrapper}>
                {!isRolesChanged && (
                    <ImageVersionPicker
                        containerId={container.id ?? ""}
                        image={container.imageName ?? ""}
                        value={selectedTag}
                        onChange={setSelectedTag}
                    />
                )}
                <Input label="Service name" inputValue={serviceName} onChange={setServiceName}/>
                {isPgLoginRequired && (
                    <PgLoginFields
                        superuser={superuser}
                        password={password}
                        onSuperuserChange={setSuperuser}
                        onPasswordChange={setPassword}
                    />
                )}
                {isRegistryLoginRequired && (
                    <RegistryLoginFields
                        username={registryUsername}
                        password={registryPassword}
                        onUsernameChange={setRegistryUsername}
                        onPasswordChange={setRegistryPassword}
                    />
                )}
                {runner && (
                    <RunnerFields
                        form={runner}
                        isRegistrationTokenFound={isRegistrationTokenFound}
                        onChange={setRunner}
                    />
                )}
                {isLinkingShown && links.length > 0 && (
                    <BindMountLinks links={links} onVolumeNameChange={handleVolumeNameChange}/>
                )}
                {isPortsEditable && (
                    <PortMappingTable
                        rows={portRows}
                        hasVolumes={hasVolumes(container)}
                        isInvalid={isRowsInvalid}
                        onRowsChange={setPortRows}
                    />
                )}
                {!isRolesChanged && !hasPublishedPorts && <NoPortsNotice/>}
                <RestartNotice isClusterMode={!isRecreating}/>
            </div>
        </div>
    )
}
