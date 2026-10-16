import type {DockerContainer} from "@/app/api/velez"
import cls from "@/pages/container/components/ContainerDetailsTab/ContainerDetailsTab.module.css"
import {formatContainerTimestamp} from "@/processes/mappings/containers.ts"
import ContainerInfoBox from "@/pages/container/components/ContainerInfoBox/ContainerInfoBox.tsx"
import ContainerMetaRow from "@/pages/container/components/ContainerMetaRow/ContainerMetaRow.tsx"

interface Props {
    container: DockerContainer
}

export default function ContainerDetailsTab({container}: Props) {
    return (
        <div className={cls.ContainerDetailsTabContainer}>
            <ContainerInfoBox title="Details">
                {container.imageName && <ContainerMetaRow label="Image">{container.imageName}</ContainerMetaRow>}
                {container.id && <ContainerMetaRow label="Id">{container.id}</ContainerMetaRow>}
                {container.createdAt && (
                    <ContainerMetaRow label="Created at">
                        {formatContainerTimestamp(container.createdAt)}
                    </ContainerMetaRow>
                )}
            </ContainerInfoBox>

            {(container.env && Object.keys(container.env).length > 0) && (
                <ContainerInfoBox title="Environment">
                    {Object.entries(container.env).map(function renderEnv([key, value]) {
                        return <ContainerMetaRow key={key} label={key}>{value}</ContainerMetaRow>
                    })}
                </ContainerInfoBox>
            )}

            {(container.networks && container.networks.length > 0) && (
                <ContainerInfoBox title="Networks">
                    {container.networks.map(function renderNetwork(n, i) {
                        const aliases = n.aliases && n.aliases.length > 0
                            ? ` (aliases: ${n.aliases.join(", ")})`
                            : ""
                        return (
                            <ContainerMetaRow key={i} label={n.networkName || ""}>
                                {n.ipAddress || "—"}{aliases}
                            </ContainerMetaRow>
                        )
                    })}
                </ContainerInfoBox>
            )}

            {(container.mounts && container.mounts.length > 0) && (
                <ContainerInfoBox title="Mounts">
                    {container.mounts.map(function renderMount(m, i) {
                        const readOnly = m.readWrite ? "" : " (ro)"
                        return (
                            <ContainerMetaRow key={i} label={m.type || "mount"}>
                                {m.source} → {m.destination}{readOnly}
                            </ContainerMetaRow>
                        )
                    })}
                </ContainerInfoBox>
            )}

            {(container.ports && container.ports.length > 0) && (
                <ContainerInfoBox title="Ports">
                    {container.ports.map(function renderPort(p, i) {
                        return (
                            <ContainerMetaRow key={i} label="Port">
                                {p.servicePortNumber} → {p.exposedTo ?? "not exposed"} ({p.protocol})
                            </ContainerMetaRow>
                        )
                    })}
                </ContainerInfoBox>
            )}
        </div>
    )
}
