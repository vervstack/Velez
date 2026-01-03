import cls from "@/components/base/Checkbox.module.css";
import { FocusEventHandler, useState} from "react";

interface CheckboxProps {
    label?: string;
    onChange: (v: boolean) => void;
    checked: boolean;
}

export default function Checkbox({label, onChange, checked}: CheckboxProps) {
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
                onChange={(e) => {
                    onChange(e.target.checked)
                }}
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
