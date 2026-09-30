import {useEffect, useMemo, useState} from "react"

import cls from "@/dialogs/CreateServiceDialog/components/AdoptForm/AdoptForm.module.css"
import type {DockerContainer, RegisterContainerRequest} from "@/app/api/velez"
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import {useEnvironmentStore} from "@/app/hooks/environment/Environment.ts"
import {IsStatefullModeEnabled} from "@/processes/queries/control_plane.ts"
import Input from "@/components/base/Input.tsx"
import AdoptActions from "@/dialogs/CreateServiceDialog/components/AdoptActions/AdoptActions.tsx"
import BindMountLinks from "@/dialogs/CreateServiceDialog/components/BindMountLinks/BindMountLinks.tsx"
import PgLoginFields from "@/dialogs/CreateServiceDialog/components/PgLoginFields/PgLoginFields.tsx"
import PortOptions from "@/dialogs/CreateServiceDialog/components/PortOptions/PortOptions.tsx"
import RegisterProgress from "@/dialogs/CreateServiceDialog/components/RegisterProgress/RegisterProgress.tsx"
import RegistryLoginFields from "@/dialogs/CreateServiceDialog/components/RegistryLoginFields/RegistryLoginFields.tsx"
import RestartConfirm from "@/dialogs/CreateServiceDialog/components/RestartConfirm/RestartConfirm.tsx"
import RestartNotice from "@/dialogs/CreateServiceDialog/components/RestartNotice/RestartNotice.tsx"
import {bindMountsOf, resolveLinks} from "@/dialogs/CreateServiceDialog/processes/bindMounts.ts"
import {
    buildRegisterContainerRequest,
    RegisterPattern,
} from "@/dialogs/CreateServiceDialog/processes/buildRegisterContainerRequest.ts"
import {parsePortRows, PortRow, publishedPortsOf} from "@/dialogs/CreateServiceDialog/processes/portRows.ts"

interface Props {
    container: DockerContainer
    pattern: RegisterPattern
    isPgLoginRequired?: boolean
    isRegistryLoginRequired?: boolean
    onBusyChange(isBusy: boolean): void
}

export default function AdoptForm({
    container,
    pattern,
    isPgLoginRequired = false,
    isRegistryLoginRequired = false,
    onBusyChange,
}: Props) {
    const [serviceName, setServiceName] = useState(container.name ?? "")
    const [linkOverrides, setLinkOverrides] = useState<Record<string, string>>({})
    const [isKeepingPorts, setIsKeepingPorts] = useState(false)
    const [portRows, setPortRows] = useState<PortRow[]>([])
    const [superuser, setSuperuser] = useState("")
    const [password, setPassword] = useState("")
    const [registryUsername, setRegistryUsername] = useState("")
    const [registryPassword, setRegistryPassword] = useState("")
    const [isConfirming, setIsConfirming] = useState(false)
    const [submittedReq, setSubmittedReq] = useState<RegisterContainerRequest | null>(null)

    const {CloseDialog} = useDialog()
    const environment = useEnvironmentStore((state) => state.selectedEnvironment)
    const isClusterMode = IsStatefullModeEnabled()

    useEffect(() => {
        onBusyChange(submittedReq !== null)
    }, [submittedReq])

    const publishedPorts = useMemo(() => publishedPortsOf(container), [container])
    const links = useMemo(
        () => resolveLinks(bindMountsOf(container), serviceName, linkOverrides),
        [container, serviceName, linkOverrides],
    )
    const parsedPorts = useMemo(() => parsePortRows(portRows), [portRows])

    const isPortsEditable = !isClusterMode && publishedPorts.length > 0
    const isKeepingEditablePorts = isPortsEditable && isKeepingPorts
    const isRowsInvalid = isPortsEditable && !isKeepingPorts && parsedPorts === null

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
            links,
            isClusterMode,
            isKeepingPorts: isKeepingEditablePorts,
            ports: isPortsEditable ? (parsedPorts ?? []) : [],
        })
    }, [container.id, environment, serviceName, pattern, isPgLoginRequired, superuser, password,
        isRegistryLoginRequired, registryUsername, registryPassword, links, isClusterMode, isKeepingEditablePorts, isPortsEditable, parsedPorts, isRowsInvalid])

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
                isKeepingPorts={isKeepingEditablePorts}
                onConfirm={handleConfirm}
                onCancel={handleCancelConfirm}
            />
        )
    }

    return (
        <div className={cls.AdoptFormContainer}>
            <div className={cls.FieldsWrapper}>
                <Input label="Image" inputValue={container.imageName ?? ""}/>
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
                {links.length > 0 && <BindMountLinks links={links} onVolumeNameChange={handleVolumeNameChange}/>}
                {isPortsEditable && (
                    <PortOptions
                        publishedPorts={publishedPorts}
                        isKeepingPorts={isKeepingPorts}
                        rows={portRows}
                        isRowsInvalid={isRowsInvalid}
                        onKeepChange={setIsKeepingPorts}
                        onRowsChange={setPortRows}
                    />
                )}
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
