import cls from "@/components/base/Input.module.css";
import {FocusEventHandler, useState} from "react";
import cn from "classnames";

export interface StyleProps {
    borderless?: boolean
}

export interface InputProps {
    label?: string;
    onChange: (v: string) => void;
    inputValue: string | null;

    onLeave?: (val: string) => void

    style?: StyleProps
}

export default function Input({label, onChange, inputValue, style, onLeave}: InputProps) {
    const [isFocused, setIsFocused] = useState(false);


    const hasValue = inputValue !== undefined && inputValue !== null && inputValue.toString().length > 0;
    const showFloatingLabel = isFocused || hasValue;

    const handleFocus: FocusEventHandler<HTMLInputElement> = () => {
        setIsFocused(true);
    };

    const handleBlur: FocusEventHandler<HTMLInputElement> = () => {
        setIsFocused(false);
        if (onLeave) onLeave(inputValue || '')
    };

    return (
        <div
            className={cn(cls.InputContainer, {
                [cls.Borderless]: style?.borderless
            })}
        >
            <input
                disabled={onChange === undefined && onLeave == undefined}
                onChange={(e) => {
                    if (onChange) onChange(e.target.value)
                }}
                onFocus={handleFocus}
                onBlur={handleBlur}
                value={inputValue || ''}
            />
            {label && (
                <label className={`${cls.Label} ${showFloatingLabel ? cls.Floating : ''}`}>
                    {label}
                </label>
            )}
        </div>
    );
}
