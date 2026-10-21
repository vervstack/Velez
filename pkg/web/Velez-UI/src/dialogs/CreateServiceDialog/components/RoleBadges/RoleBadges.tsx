import cls from "@/dialogs/CreateServiceDialog/components/RoleBadges/RoleBadges.module.css"
import Badge from "@/components/base/Badge.tsx"

interface Props {
    isService: boolean
    isRoot: boolean
}

export default function RoleBadges({isService, isRoot}: Props) {
    return (
        <span className={cls.RoleBadgesContainer}>
            {isService && (
                <span className={cls.Enter}>
                    <Badge label="Service" color="var(--green)" dim="var(--green-dim)"/>
                </span>
            )}
            {isRoot && (
                <span className={cls.Enter}>
                    <Badge label="Network root" color="var(--blue)" dim="var(--blue-dim)"/>
                </span>
            )}
        </span>
    )
}
