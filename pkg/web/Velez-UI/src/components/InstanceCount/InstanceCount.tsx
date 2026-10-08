import cls from "@/components/InstanceCount/InstanceCount.module.css"

interface Props {
    count: number
    label: string
    isLoading: boolean
}

export default function InstanceCount({count, label, isLoading}: Props) {
    if (isLoading) {
        return (
            <span className={cls.InstanceCountContainer} aria-busy="true">
                <span className={cls.Skeleton}/>
                {label}
            </span>
        )
    }

    return <span className={cls.InstanceCountContainer}>{count} {label}</span>
}
