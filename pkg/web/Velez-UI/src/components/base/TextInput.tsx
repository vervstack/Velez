import cls from '@/components/base/TextInput.module.css';
import cn from 'classnames';

interface TextInputProps {
    value: string;
    onChange?: (v: string) => void;
    prefix?: React.ReactNode;
    suffix?: React.ReactNode;
    mono?: boolean;
    readOnly?: boolean;
    placeholder?: string;
    disabled?: boolean;
}

export default function TextInput({
    value,
    onChange,
    prefix,
    suffix,
    mono,
    readOnly,
    placeholder,
    disabled,
}: TextInputProps) {
    return (
        <div className={cn(cls.TextInputContainer, { [cls.readOnly]: readOnly })}>
            {prefix && <span className={cn(cls.Chip, cls.prefix)}>{prefix}</span>}
            <input
                className={cn(cls.input, { [cls.mono]: mono })}
                value={value}
                onChange={(e) => onChange?.(e.target.value)}
                readOnly={readOnly || !onChange}
                disabled={disabled}
                placeholder={placeholder}
            />
            {suffix && <span className={cn(cls.Chip, cls.suffix)}>{suffix}</span>}
        </div>
    );
}
