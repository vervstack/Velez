import cls from "@/components/base/ActionButton.module.css";

interface ActionButtonProps {
    iconPath: string
    alt?: string
    onClick?: () => void
}

export default function ActionButton({iconPath, alt, onClick}: ActionButtonProps) {
    return (
        <div className={cls.AddButton}
             onClick={onClick}>
            <img src={iconPath}
                 alt={alt || iconPath}/>
        </div>
    )
}
