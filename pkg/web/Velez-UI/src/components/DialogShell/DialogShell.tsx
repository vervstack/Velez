import {ReactNode} from "react"
import cn from "classnames"

import cls from "@/components/DialogShell/DialogShell.module.css"
import Button from "@/components/base/Button.tsx"

interface Props {
    title: string
    onClose?: () => void
    isFlush?: boolean
    children: ReactNode
}

export default function DialogShell({title, onClose, isFlush, children}: Props) {
    return (
        <div className={cls.DialogShellContainer}>
            <div className={cls.Header}>
                <h2 className={cls.Title}>{title}</h2>
                {onClose && <Button variant="ghost" sm onClick={onClose}>✕</Button>}
            </div>
            <div className={cn(cls.Body, isFlush && cls.Flush)}>
                {children}
            </div>
        </div>
    )
}
