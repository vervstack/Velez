import Button from "@/components/base/Button.tsx"
import cls from "@/pages/runners/components/RunnersEmptyState/RunnersEmptyState.module.css"

interface Props {
    onCreate: () => void
}

export default function RunnersEmptyState({onCreate}: Props) {
    return (
        <div className={cls.RunnersEmptyStateContainer}>
            <span className={cls.Icon}>⚙</span>
            <div className={cls.Message}>No runners on this node.</div>
            <Button variant="primary" onClick={onCreate}>
                Create your first runner
            </Button>
        </div>
    )
}
