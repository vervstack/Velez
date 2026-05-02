import { useState, useEffect, useRef } from 'react';
import cls from '@/components/complex/ThreeDotMenu/ThreeDotMenu.module.css';
import cn from 'classnames';

export interface MenuAction {
    label: string;
    icon: string;
    danger?: boolean;
    onClick?: () => void;
}

interface ThreeDotMenuProps {
    actions: MenuAction[];
}

export default function ThreeDotMenu({ actions }: ThreeDotMenuProps) {
    const [open, setOpen] = useState(false);
    const ref = useRef<HTMLDivElement>(null);

    useEffect(function bindOutsideClick() {
        if (!open) return;

        function handleClickOutside(e: MouseEvent) {
            if (ref.current && !ref.current.contains(e.target as Node)) {
                setOpen(false);
            }
        }

        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, [open]);

    function handleTriggerClick(e: React.MouseEvent) {
        e.stopPropagation();
        setOpen(prev => !prev);
    }

    return (
        <div ref={ref} className={cls.ThreeDotMenuContainer}>
            <button
                className={cn(cls.trigger, { [cls.triggerOpen]: open })}
                onClick={handleTriggerClick}
                aria-label="Open menu"
            >
                ···
            </button>

            {open && (
                <div className={cls.dropdown}>
                    {actions.map(function renderAction(action, i) {
                        function handleActionClick() {
                            setOpen(false);
                            action.onClick?.();
                        }

                        return (
                            <button
                                key={i}
                                className={cn(cls.action, { [cls.actionDanger]: action.danger })}
                                onClick={handleActionClick}
                            >
                                <span className={cls.actionIcon}>{action.icon}</span>
                                {action.label}
                            </button>
                        );
                    })}
                </div>
            )}
        </div>
    );
}
