import cn from "classnames";

import cls from "@/components/complex/dialog/Dialog.module.css";

interface DialogProps {
    isOpen: boolean;
    closeDialog: () => void;

    children: React.ReactNode;
}

export default function Dialog({isOpen, children, closeDialog}: DialogProps) {
    return (
        <div className={cn(cls.DialogContainer, {
            [cls.isOpen]: isOpen,
        })}>
            <div className={cls.Content}>
                {children}
            </div>
            <div
                className={cls.CloseButton}
                onClick={ closeDialog}
            >X</div>
        </div>
    )
}
