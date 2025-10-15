import cls from "@/components/base/Input.module.css";
import {FocusEventHandler, useEffect, useState} from "react";
import cn from "classnames";

export interface StyleProps {
    borderless?: boolean
}

interface InputProps {
    label?: string;
    onChange?: (v: string) => void;
    onLeave?: (val: string) => void
    inputValue?: string | null;

    style?: StyleProps
}

export default function Input({label, onChange, inputValue, style, onLeave}: InputProps) {
    const [isFocused, setIsFocused] = useState(false);

    const [value, setValue] = useState(inputValue || '');

    useEffect(() => {
        if (onChange) onChange(value)
    }, [value]);

    const hasValue = value !== undefined && value !== null && value.toString().length > 0;
    const showFloatingLabel = isFocused || hasValue;

    const handleFocus: FocusEventHandler<HTMLInputElement> = () => {
        setIsFocused(true);
    };

    const handleBlur: FocusEventHandler<HTMLInputElement> = () => {
        setIsFocused(false);
        if (onLeave) onLeave(value)
    };

    return (
        <div
            className={cn(cls.InputContainer, {
                [cls.Borderless]: style?.borderless
            })}
        >
            <input
                onChange={(e) => setValue(e.target.value)}
                onFocus={handleFocus}
                onBlur={handleBlur}
                value={value}
            />
            {label && (
                <label className={`${cls.Label} ${showFloatingLabel ? cls.Floating : ''}`}>
                    {label}
                </label>
            )}
        </div>
    );
}
