import cn from "classnames";

import {motion, AnimatePresence} from "framer-motion";
import {useEffect, useState} from "react";

import cls from '@/segments/Toast.module.css';

import {Toast as ToastProp, useToaster} from "@/app/hooks/toaster/Toaster.ts";

export default function Toaster() {
    const {toasts} = useToaster();

    return (
        <div className={cls.ToastContainer}>
            <AnimatePresence>
                {toasts.map((toast, idx) => (
                    <motion.div
                        key={idx}
                        layout
                        initial={{opacity: 0, x: 100, y: 50}} // slide in from right + slight bottom offset
                        animate={{opacity: 1, x: 0, y: 0}}
                        exit={{opacity: 0, y: -50}} // move up and fade out
                        transition={{type: "spring", stiffness: 300, damping: 20}}

                    >
                        <Toast {...toast}/>
                    </motion.div>
                ))}
            </AnimatePresence>
        </div>
    )
}


function Toast({title, description, isDismissable}: ToastProp) {
    const [isLeaving, setIsLeaving] = useState(false);
    const toaster = useToaster();

    useEffect(() => {
        return () => {
            setIsLeaving(true);
        };
    }, []);

    return (
        <div
            className={cn(cls.Toast, {
                [cls.error]: title === 'Error',
                [cls.warn]: title === 'Warn',
                [cls.info]: title == undefined || title === 'Info',

                [cls.slideIn]: !isLeaving,
                [cls.slideOut]: isLeaving,
            })}>
            <div>{title}</div>
            <div className={cls.Description}> {description}</div>

            {isDismissable && <div
				className={cls.DismissButton}
				onClick={() => toaster.dismiss(title)}
			>
				-
			</div>}
        </div>
    )
}
