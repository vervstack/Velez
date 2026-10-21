import cn from "classnames"

import cls from "@/dialogs/CreateServiceDialog/components/ContainerOverviewRow/ContainerOverviewRow.module.css"
import type {DockerContainer} from "@/app/api/velez"
import RoleSelectButton from "@/dialogs/CreateServiceDialog/components/RoleSelectButton/RoleSelectButton.tsx"

const ROOT_TOOLTIP = "The container that owns the network namespace (often a Tailscale or VPN container). "
    + "Pick one that lives long and restarts rarely: restarting it restarts every container attached to it."
const SERVICE_TOOLTIP = "The main application you want to run and watch."

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
            <div className={cls.Meta}>
                <span className={cls.Image}>{container.imageName}</span>
                <span className={cls.Network}>{`network: ${networkMode}`}</span>
            </div>
            <div className={cls.Roles}>
                <RoleSelectButton
                    label="Network root"
                    isSelected={isRoot}
                    isDisabled={isReadOnly}
                    tooltip={ROOT_TOOLTIP}
                    onSelect={handlePickRoot}
                />
                <RoleSelectButton
                    label="Service"
                    isSelected={isService}
                    isDisabled={isReadOnly}
                    tooltip={SERVICE_TOOLTIP}
                    onSelect={handlePickService}
                />
            </div>
        </div>
    )
}
