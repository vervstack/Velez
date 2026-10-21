import cn from "classnames"

import cls from "@/dialogs/CreateServiceDialog/components/ContainerOverviewRow/ContainerOverviewRow.module.css"
import type {DockerContainer} from "@/app/api/velez"
import RoleBadges from "@/dialogs/CreateServiceDialog/components/RoleBadges/RoleBadges.tsx"
import RoleSelectButton from "@/dialogs/CreateServiceDialog/components/RoleSelectButton/RoleSelectButton.tsx"

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
            <div className={cls.Badges}>
                <RoleBadges isService={isService} isRoot={isRoot}/>
            </div>
            <div className={cls.Meta}>
                <span className={cls.Image}>{container.imageName}</span>
                <span className={cls.Network}>{`network: ${networkMode}`}</span>
            </div>
            <div className={cls.Roles}>
                {!isRoot && (
                    <RoleSelectButton
                        label="Make root container"
                        isDisabled={isReadOnly}
                        tooltip={ROOT_TOOLTIP}
                        onSelect={handlePickRoot}
                    />
                )}
                {!isService && (
                    <RoleSelectButton
                        label="Make service container"
                        isDisabled={isReadOnly}
                        tooltip={SERVICE_TOOLTIP}
                        onSelect={handlePickService}
                    />
                )}
            </div>
        </div>
    )
}
