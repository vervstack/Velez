import {useEffect, useMemo, useState} from "react"

import cls from "@/dialogs/CreateServiceDialog/components/AdoptForm/AdoptForm.module.css"
import type {DockerContainer, RegisterContainerRequest} from "@/app/api/velez"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useEnvironmentStore} from "@/app/hooks/environment/Environment.ts"
import {IsStatefullModeEnabled} from "@/processes/queries/control_plane.ts"
import Input from "@/components/base/Input.tsx"
import AdoptActions from "@/dialogs/CreateServiceDialog/components/AdoptActions/AdoptActions.tsx"
import BindMountLinks from "@/dialogs/CreateServiceDialog/components/BindMountLinks/BindMountLinks.tsx"
import ImageVersionPicker from "@/dialogs/CreateServiceDialog/components/ImageVersionPicker/ImageVersionPicker.tsx"
import NoPortsNotice from "@/dialogs/CreateServiceDialog/components/NoPortsNotice/NoPortsNotice.tsx"
import PgLoginFields from "@/dialogs/CreateServiceDialog/components/PgLoginFields/PgLoginFields.tsx"
import PortMappingTable from "@/dialogs/CreateServiceDialog/components/PortMappingTable/PortMappingTable.tsx"
import RegisterProgress from "@/dialogs/CreateServiceDialog/components/RegisterProgress/RegisterProgress.tsx"
import RegistryLoginFields from "@/dialogs/CreateServiceDialog/components/RegistryLoginFields/RegistryLoginFields.tsx"
import RestartConfirm from "@/dialogs/CreateServiceDialog/components/RestartConfirm/RestartConfirm.tsx"
import RestartNotice from "@/dialogs/CreateServiceDialog/components/RestartNotice/RestartNotice.tsx"
import RunnerFields from "@/dialogs/CreateServiceDialog/components/RunnerFields/RunnerFields.tsx"
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
    onBusyChange(isBusy: boolean): void
}

export default function AdoptForm({
    container,
    pattern,
    isPgLoginRequired = false,
    isRegistryLoginRequired = false,
    initialRunner,
    isRegistrationTokenFound = false,
    onBusyChange,
}: Props) {
    const [serviceName, setServiceName] = useState(container.name ?? "")
    const [linkOverrides, setLinkOverrides] = useState<Record<string, string>>({})
    const [portRows, setPortRows] = useState<PortMappingRow[]>(() => portMappingRowsOf(container))
    const [superuser, setSuperuser] = useState("")
    const [password, setPassword] = useState("")
    const [registryUsername, setRegistryUsername] = useState("")
    const [registryPassword, setRegistryPassword] = useState("")
    const [runner, setRunner] = useState<RunnerForm | undefined>(initialRunner)
    const [selectedTag, setSelectedTag] = useState<string | undefined>(undefined)
    const [isConfirming, setIsConfirming] = useState(false)
    const [submittedReq, setSubmittedReq] = useState<RegisterContainerRequest | null>(null)

    const {CloseDialog} = useDialog()
    const environment = useEnvironmentStore((state) => state.selectedEnvironment)
    const isClusterMode = IsStatefullModeEnabled()

    useEffect(() => {
        onBusyChange(submittedReq !== null)
    }, [submittedReq])

    const links = useMemo(
        () => resolveLinks(bindMountsOf(container), serviceName, linkOverrides),
        [container, serviceName, linkOverrides],
    )
    const parsedPorts = useMemo(() => parsePortMappingRows(portRows), [portRows])

    const hasPublishedPorts = portRows.length > 0
    const isPortsEditable = !isClusterMode && hasPublishedPorts
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
        })
    }, [container.id, environment, serviceName, pattern, isPgLoginRequired, superuser, password,
        isRegistryLoginRequired, registryUsername, registryPassword, runner, links, isClusterMode,
        isPortsEditable, parsedPorts, isRowsInvalid, selectedTag, container.imageName])

    function handleVolumeNameChange(destination: string, volumeName: string) {
        setLinkOverrides({...linkOverrides, [destination]: volumeName})
    }

    function handleConfirm() {
        setSubmittedReq(request)
    }

    function handleRegister() {
        if (isClusterMode) handleConfirm()
        else setIsConfirming(true)
    }

    function handleCancelConfirm() {
        setIsConfirming(false)
    }

    if (submittedReq) {
        return <RegisterProgress request={submittedReq} containerName={container.name}/>
    }

    if (isConfirming) {
        return (
            <RestartConfirm
                containerName={container.name}
                hasPublishedPorts={hasPublishedPorts}
                onConfirm={handleConfirm}
                onCancel={handleCancelConfirm}
            />
        )
    }

    return (
        <div className={cls.AdoptFormContainer}>
            <div className={cls.FieldsWrapper}>
                <ImageVersionPicker
                    containerId={container.id ?? ""}
                    image={container.imageName ?? ""}
                    value={selectedTag}
                    onChange={setSelectedTag}
                />
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
                {links.length > 0 && <BindMountLinks links={links} onVolumeNameChange={handleVolumeNameChange}/>}
                {isPortsEditable && (
                    <PortMappingTable
                        rows={portRows}
                        hasVolumes={hasVolumes(container)}
                        isInvalid={isRowsInvalid}
                        onRowsChange={setPortRows}
                    />
                )}
                {!hasPublishedPorts && <NoPortsNotice/>}
                <RestartNotice isClusterMode={isClusterMode}/>
            </div>
            <AdoptActions
                isClusterMode={isClusterMode}
                isRegisterDisabled={!request}
                onCancel={CloseDialog}
                onRegister={handleRegister}
            />
        </div>
    )
}
