import cls from "@/components/base/Toggle.module.css";
import cn from "classnames";

interface ToggleProps {
    value: boolean
    onChange: (value: boolean) => void
}

export default function Toggle({value, onChange}: ToggleProps) {
    return (
        <div className={
            cn(cls.ToggleContainer, {
                [cls.on]: value
            })
        }
             onClick={() => onChange(!value)}
        >
            <div className={cls.SwitchHandle}/>
        </div>
    )
}
