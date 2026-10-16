import type {DockerContainer} from "@/app/api/velez"
import cls from "@/pages/container/components/ContainerInspectTab/ContainerInspectTab.module.css"

interface Props {
    container: DockerContainer
}

export default function ContainerInspectTab({container}: Props) {
    return (
        <div className={cls.ContainerInspectTabContainer}>
            <pre className={cls.Body}>{JSON.stringify(container, null, 2)}</pre>
        </div>
    )
}
