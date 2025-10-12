import cls from "@/components/base/Checkbox.module.css";
import { ChangeEventHandler, FocusEventHandler, useState } from "react";

interface CheckboxProps {
    label?: string;
    onChange: ChangeEventHandler<HTMLInputElement> | undefined;
    checked: boolean;
}

export default function Checkbox({ label, onChange, checked }: CheckboxProps) {
    const [isFocused, setIsFocused] = useState(false);
    const showFloatingLabel = isFocused || checked;

    const handleFocus: FocusEventHandler<HTMLInputElement> = () => {
        setIsFocused(true);
    };

    const handleBlur: FocusEventHandler<HTMLInputElement> = () => {
        setIsFocused(false);
    };

    return (
        <div className={cls.CheckboxContainer}>
            <input
                type="checkbox"
                onChange={onChange}
                onFocus={handleFocus}
                onBlur={handleBlur}
                checked={checked}
            />
            {label && (
                <label className={`${cls.Label} ${showFloatingLabel ? cls.Floating : ''}`}>
                    {label}
                </label>
            )}
        </div>
    );
}
