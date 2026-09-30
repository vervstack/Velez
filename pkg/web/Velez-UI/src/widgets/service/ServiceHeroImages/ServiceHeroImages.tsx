import {useNavigate} from "react-router-dom"

import cls from "@/widgets/service/ServiceHeroImages/ServiceHeroImages.module.css"

import {Routes} from "@/app/router/Routes"
import type {ServiceSidecarView} from "@/model/service_page/ServicePageModel"

import ContainerImageTile from "@/components/smerd/ContainerImageTile/ContainerImageTile.tsx"

interface Props {
    serviceName: string
    imageName?: string
    containerId?: string
    sidecars: ServiceSidecarView[]
}

export default function ServiceHeroImages({serviceName, imageName, containerId, sidecars}: Props) {
    const navigate = useNavigate()

    function handleOpen(id: string) {
        navigate(Routes.Container + "/" + id)
    }

    const hasMain = !!imageName && !!containerId
    if (!hasMain && sidecars.length === 0) return null

    return (
        <div className={cls.ServiceHeroImagesContainer}>
            {hasMain && (
                <ContainerImageTile
                    imageName={imageName}
                    label={serviceName}
                    onOpen={() => handleOpen(containerId)}
                />
            )}
            {sidecars.map(sidecar => (
                <ContainerImageTile
                    key={sidecar.containerId}
                    imageName={sidecar.imageName}
                    label={sidecar.name}
                    status={sidecar.status}
                    onOpen={() => handleOpen(sidecar.containerId)}
                />
            ))}
        </div>
    )
}
