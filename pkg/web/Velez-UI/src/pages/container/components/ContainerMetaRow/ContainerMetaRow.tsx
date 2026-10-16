import cls from "@/pages/container/components/ContainerMetaRow/ContainerMetaRow.module.css"

interface Props {
    label: string
    children: React.ReactNode
}

export default function ContainerMetaRow({label, children}: Props) {
    return (
        <div className={cls.ContainerMetaRowContainer}>
            <span className={cls.Label}>{label}</span>
            <span className={cls.Value}>{children}</span>
        </div>
    )
}
