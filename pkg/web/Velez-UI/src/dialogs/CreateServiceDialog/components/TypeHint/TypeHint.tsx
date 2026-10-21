import cn from "classnames"

import cls from "@/dialogs/CreateServiceDialog/components/TypeHint/TypeHint.module.css"

interface Props {
    isVisible: boolean
    text: string
}

export default function TypeHint({isVisible, text}: Props) {
    return (
        <div
            role="status"
            aria-hidden={!isVisible}
            className={cn(cls.TypeHintContainer, isVisible && cls.Visible)}
        >
            <div className={cls.Inner}>
                <span className={cls.Text}>{text}</span>
            </div>
        </div>
    )
}
