import {MouseEvent} from "react";

import cls from "@/pages/services/parts/ContainerCard/ContainerCard.module.css";
import type {DockerContainer} from "@/app/api/velez";
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx";
import {
    deriveContainerServicePresentation,
    mapContainerStatusToDot,
    containerStatusLabel,
    suggestedPatternHint,
} from "@/processes/mappings/containers";
import Button from "@/components/base/Button.tsx";
import StatusDot from "@/components/base/StatusDot";
import ServiceChip from "@/components/base/chips/ServiceChip";
import CreateServiceDialog from "@/dialogs/CreateServiceDialog/CreateServiceDialog.tsx";

interface Props {
    container: DockerContainer;
    onOpen: (id: string) => void;
    onFilterByService: (serviceName: string) => void;
    isRegisterHidden?: boolean;
}

export default function ContainerCard({container, onOpen, onFilterByService, isRegisterHidden = false}: Props) {
    const {OpenDialog} = useDialog();

    function handleClick() {
        onOpen(container.id || "");
    }

    function handleActionsClick(e: MouseEvent) {
        e.stopPropagation();
    }

    function handleRegister() {
        OpenDialog(<CreateServiceDialog container={container}/>);
    }

    const presentation = container.linkedServiceName
        ? deriveContainerServicePresentation(container)
        : null;
    const hint = suggestedPatternHint(container.suggestedPattern);

    return (
        <div className={cls.ContainerCardContainer} onClick={handleClick}>
            <div className={cls.NameRow}>
                <span className={cls.Name}>{container.name || container.id}</span>
                {!container.isRegistered && <span className={cls.UnregisteredBadge}>unregistered</span>}
            </div>
            <div className={cls.Image}>{container.imageName}</div>
            <div className={cls.StatusRow}>
                <StatusDot
                    status={mapContainerStatusToDot(container.status)}
                    tooltip={containerStatusLabel(container.status)}
                />
                {presentation && (
                    <ServiceChip
                        serviceName={presentation.displayName}
                        icon={presentation.icon}
                        onClick={onFilterByService}
                    />
                )}
            </div>
            {!container.linkedServiceName && !isRegisterHidden && (
                <div className={cls.ActionsRow} onClick={handleActionsClick}>
                    {hint && <span className={cls.Hint}>{hint}</span>}
                    <Button variant="secondary" sm onClick={handleRegister}>Register</Button>
                </div>
            )}
        </div>
    );
}
