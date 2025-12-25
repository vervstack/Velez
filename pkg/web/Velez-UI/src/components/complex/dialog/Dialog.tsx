import cn from "classnames";

import cls from "@/components/complex/dialog/Dialog.module.css";

interface DialogProps {
    isOpen: boolean;
    closeDialog: () => void;

    children: React.ReactNode;

    blur?: number; // float from 0 to 1
}

export default function Dialog({isOpen, children, closeDialog, blur}: DialogProps) {
    return (
        <>
            <div
                className={cn(cls.Backdrop, {
                    [cls.isOpen]: isOpen,
                })}
                onClick={closeDialog}
                style={{backdropFilter: `blur(${(blur || 0) * 10}px)`}}
            />

            <div className={cn(cls.DialogContainer, {
                [cls.isOpen]: isOpen,
            })}>
                <div className={cls.Content}>
                    {children}
                </div>
                <div
                    className={cls.CloseButton}
                    onClick={closeDialog}
                >X
                </div>
            </div>
        </>
    )
}
