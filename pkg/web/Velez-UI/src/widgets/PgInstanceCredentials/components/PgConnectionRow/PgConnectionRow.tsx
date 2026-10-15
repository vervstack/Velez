import cls from "@/widgets/PgInstanceCredentials/components/PgConnectionRow/PgConnectionRow.module.css"
import Button from "@/components/base/Button.tsx"

interface Props {
    label: string
    maskedDsn?: string
    isLoading: boolean
    onCopy: () => void
}

export default function PgConnectionRow({label, maskedDsn, isLoading, onCopy}: Props) {
    return (
        <div className={cls.PgConnectionRowContainer}>
            <span className={cls.Label}>{label}</span>
            <span className={cls.Value}>{isLoading ? "loading…" : maskedDsn || "-"}</span>
            <Button sm disabled={isLoading || !maskedDsn} onClick={onCopy}>
                Copy
            </Button>
        </div>
    )
}
