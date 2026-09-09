import Button from "@/components/base/Button.tsx"
import cls from "@/pages/postgres/components/PostgresEmptyState/PostgresEmptyState.module.css"

interface Props {
    onCreate: () => void
}

export default function PostgresEmptyState({onCreate}: Props) {
    return (
        <div className={cls.PostgresEmptyStateContainer}>
            <span className={cls.Icon}>⛁</span>
            <div className={cls.Message}>No Postgres instances on this node.</div>
            <Button variant="primary" onClick={onCreate}>
                Create your first database
            </Button>
        </div>
    )
}
