import cls from '@/components/base/IconButton.module.css';
import cn from 'classnames';

interface IconButtonProps {
    label: string;
    title?: string;
    onClick?: () => void;
    danger?: boolean;
    disabled?: boolean;
}

export default function IconButton({ label, title, onClick, danger, disabled }: IconButtonProps) {
    return (
        <button
            className={cn(cls.IconButtonContainer, { [cls.danger]: danger })}
            title={title}
            onClick={onClick}
            disabled={disabled}
        >
            {label}
        </button>
    );
}
