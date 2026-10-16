import cls from "@/pages/container/components/ContainerInfoBox/ContainerInfoBox.module.css"

interface Props {
    title: string
    children: React.ReactNode
}

export default function ContainerInfoBox({title, children}: Props) {
    return (
        <div className={cls.ContainerInfoBoxContainer}>
            <div className={cls.Title}>{title}</div>
            <div className={cls.Body}>{children}</div>
        </div>
    )
}
