import cn from "classnames";

import cls from "@/components/complex/dialog/Dialog.module.css";
import {useEffect, useState} from "react";

interface DialogProps {
    isOpen: boolean;
    onClose: () => void;

    children: React.ReactNode;

    blur?: number; // float from 0 to 1
}

export default function Dialog({isOpen, children, onClose, blur}: DialogProps) {
    const [open, setOpen] = useState(isOpen);

    function close() {
        setOpen(false)
        setTimeout(() => onClose(), 500)
    }

    useEffect(() => {
        setOpen(isOpen);

        function handleKeyDown(event: KeyboardEvent) {
            if (event.key === 'Escape' && isOpen) {
                close();
            }
        }

        if (isOpen) {
            window.addEventListener('keydown', handleKeyDown);
        }

        return () => {
            window.removeEventListener('keydown', handleKeyDown);
        };
    }, [isOpen, onClose])

    return (
        <>
            <div
                className={cn(cls.Backdrop, {
                    [cls.isOpen]: open,
                })}
                onClick={close}
                style={{backdropFilter: `blur(${(blur || 0) * 10}px)`}}
            />

            <div className={cn(cls.DialogContainer, {
                [cls.isOpen]: open,
            })}>
                <div className={cls.Content}>
                    {/*TODO проблема быстрого сворачивания диалога - тут. Ребёнок удоляется быстрее чем сворачивается*/}
                    {children}
                </div>
                <div
                    className={cls.CloseButton}
                    onClick={close}
                >X
                </div>
            </div>
        </>
    )
}
