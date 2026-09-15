import Button from "@/components/base/Button.tsx"
import cls from "@/pages/container-registry/components/ContainerRegistryEmptyState/ContainerRegistryEmptyState.module.css"

interface Props {
    onCreate: () => void
}

export default function ContainerRegistryEmptyState({onCreate}: Props) {
    return (
        <div className={cls.ContainerRegistryEmptyStateContainer}>
            <span className={cls.Icon}>▣</span>
            <div className={cls.Message}>No container registries on this node.</div>
            <Button variant="primary" onClick={onCreate}>
                Create your first registry
            </Button>
        </div>
    )
}
