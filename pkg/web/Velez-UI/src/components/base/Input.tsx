import {FocusEventHandler, useState} from "react";
import cn from "classnames";

import cls from "@/components/base/Input.module.css";

import InfoMark from "@/assets/icons/InfoMark.svg";

export interface StyleProps {
    borderless?: boolean
}

export interface InputProps {
    label?: string;
    inputValue: string | null;
    onChange: (v: string) => void;

    onLeave?: (val: string) => void

    style?: StyleProps

    disabled?: boolean

    hint?: string
}

export default function Input({label, onChange, inputValue, style, onLeave, disabled, hint}: InputProps) {
    const [isFocused, setIsFocused] = useState(false);

    const hasValue = inputValue !== undefined && inputValue !== null && inputValue.toString().length > 0;
    const showFloatingLabel = isFocused || hasValue;

    const handleFocus: FocusEventHandler<HTMLInputElement> = () => {
        if (disabled) {
            return
        }
        setIsFocused(true);
    };

    const handleBlur: FocusEventHandler<HTMLInputElement> = () => {
        setIsFocused(false);
        if (onLeave) onLeave(inputValue || '')
    };

    return (
        <div
            className={cn(cls.InputContainer, {
                [cls.Borderless]: style?.borderless,
                [cls.Disabled]: disabled,
            })}
        >
            <input
                className={cn(cls.input, {
                    [cls.Disabled]: disabled,
                })}

                disabled={(onChange === undefined && onLeave == undefined) || disabled}
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
            {
                hint && (
                    <img
                        className={cls.Hint}
                        src={InfoMark}
                        alt={'?'}
                        data-tooltip-id={"tooltip"}
                        data-tooltip-content={hint}
                        data-tooltip-place="top"
                    />
                )
            }
        </div>
    );
}
