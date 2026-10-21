import cn from "classnames"

import cls from "@/dialogs/CreateServiceDialog/components/ContainerOverviewRow/ContainerOverviewRow.module.css"
import type {DockerContainer} from "@/app/api/velez"
import RoleSlot from "@/dialogs/CreateServiceDialog/components/RoleSlot/RoleSlot.tsx"

const ROOT_TOOLTIP = [
    "Owns the network namespace of the group.",
    "Often a Tailscale or VPN container.",
    "Pick one that lives long and restarts rarely:",
    "restarting it restarts every attached container.",
].join("\n")
const SERVICE_TOOLTIP = [
    "The main application you want to run and watch.",
    "Usually the business logic of the group.",
].join("\n")

interface Props {
    container: DockerContainer
    networkMode: string
    isRoot: boolean
    isService: boolean
    isReadOnly: boolean
    onPickRoot(id: string): void
    onPickService(id: string): void
}

export default function ContainerOverviewRow(
    {container, networkMode, isRoot, isService, isReadOnly, onPickRoot, onPickService}: Props,
) {
    const containerId = container.id ?? ""

    function handlePickRoot() {
        onPickRoot(containerId)
    }

    function handlePickService() {
        onPickService(containerId)
    }

    return (
        <div className={cn(cls.ContainerOverviewRowContainer, (isRoot || isService) && cls.HasRole)}>
            <span className={cls.Name}>{container.name}</span>
            <div className={cls.Roles}>
                <RoleSlot
                    chipLabel="Network root"
                    chipColor="var(--blue)"
                    chipDim="var(--blue-dim)"
                    buttonLabel="Make root container"
                    isHeld={isRoot}
                    isDisabled={isReadOnly}
                    tooltip={ROOT_TOOLTIP}
                    onSelect={handlePickRoot}
                />
                <RoleSlot
                    chipLabel="Service"
                    chipColor="var(--green)"
                    chipDim="var(--green-dim)"
                    buttonLabel="Make service container"
                    isHeld={isService}
                    isDisabled={isReadOnly}
                    tooltip={SERVICE_TOOLTIP}
                    onSelect={handlePickService}
                />
            </div>
            <div className={cls.Meta}>
                <span className={cls.Image}>{container.imageName}</span>
                <span className={cls.Network}>{`network: ${networkMode}`}</span>
            </div>
        </div>
    )
}
