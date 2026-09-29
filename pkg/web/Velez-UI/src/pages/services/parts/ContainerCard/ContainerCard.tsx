import {DockerContainer} from "@/app/api/velez";
import StatusDot from "@/components/base/StatusDot";
import ServiceChip from "@/components/base/chips/ServiceChip";
import {deriveContainerServicePresentation, mapContainerStatusToDot, containerStatusLabel} from "@/processes/mappings/containers";

import cls from "@/pages/services/parts/ContainerCard/ContainerCard.module.css";

interface Props {
    container: DockerContainer;
    onOpen: (id: string) => void;
    onFilterByService: (serviceName: string) => void;
}

export default function ContainerCard({container, onOpen, onFilterByService}: Props) {
    function handleClick() {
        onOpen(container.id || "");
    }

    const presentation = container.linkedServiceName
        ? deriveContainerServicePresentation(container)
        : null;

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
        </div>
    );
}
