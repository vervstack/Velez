import cls from "@/components/base/Input.module.css";

import {ChangeEventHandler} from "react";

interface InputProps {
    onChange: ChangeEventHandler<HTMLInputElement> | undefined
    value: string | number
}

export default function Input({onChange, value}: InputProps) {
    return (
        <div className={cls.InputContainer}>
            <input
                onChange={onChange}
                value={value}
            />
        </div>
    )
}
