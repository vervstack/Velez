import cls from '@/components/base/Button.module.css';
import cn from 'classnames';

interface ButtonProps {
    children?: React.ReactNode;
    /** @deprecated use children */
    title?: string;
    variant?: 'primary' | 'secondary' | 'danger';
    onClick?: () => void;
    disabled?: boolean;
    /** @deprecated use disabled */
    isDisabled?: boolean;
    type?: 'button' | 'submit' | 'reset';
}

export default function Button({
    children,
    title,
    variant = 'secondary',
    onClick,
    disabled,
    isDisabled,
    type = 'button',
}: ButtonProps) {
    const isOff = disabled ?? isDisabled;
    return (
        <button
            className={cn(cls.ButtonContainer, cls[variant])}
            onClick={onClick}
            disabled={isOff}
            type={type}
        >
            {children ?? title}
        </button>
    );
}
