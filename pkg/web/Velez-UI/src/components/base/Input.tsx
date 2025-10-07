import cls from "@/components/base/Input.module.css";

import {ChangeEventHandler} from "react";

interface InputProps {
    label?: string

    onChange: ChangeEventHandler<HTMLInputElement> | undefined
    value: string | number
}

export default function Input({label, onChange, value}: InputProps) {
    return (
        <div className={cls.InputContainer}>
            {label ? <div>{label}</div> : null}
            <input
                onChange={onChange}
                value={value}
            />
        </div>
    )
}
