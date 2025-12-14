import cls from "@/components/base/Button.module.css";
import cn from "classnames";

interface ButtonProps {
    title: string
    isDisabled?: boolean
    onClick?: () => void
}

export default function Button({onClick, title, isDisabled}: ButtonProps) {
    return (
        <button
            className={cn(cls.ButtonContainer, {
                [cls.disabled]: isDisabled
            })}
            onClick={onClick}
            disabled={isDisabled}
        >
            {title}
        </button>
    )
}
