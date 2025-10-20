import cls from "@/components/complex/mapping/ValuesPair.module.css";

import ArrowForward from "@/assets/icons/ArrowForward.svg";
import Input from "@/components/base/Input.tsx";

interface VolumeMappingProps {
    name: string
    value?: string

    onChange?: (name: string, value?: string) => void
}

export default function ValuesPair({name, value, onChange}: VolumeMappingProps) {
    const handleNameChange = onChange ? (newName: string) => onChange(newName, value) : undefined
    const handleValueChange = onChange ? (newValue: string) => onChange(name, newValue) : undefined

    return (
        <div className={cls.ValuesPairContainer}>
            <div
                className={cls.Value}

                data-tooltip-id={"tooltip"}
                data-tooltip-content="Container path"
                data-tooltip-place="top"
            >
                <Input
                    inputValue={name}
                    onChange={handleNameChange}
                />
            </div>

            <img
                className={cls.MapArrow}
                src={ArrowForward}
                alt={'->'}
            />

            <div
                className={cls.Value}
                data-tooltip-id={"tooltip"}
                data-tooltip-content="Host path"
                data-tooltip-place="bottom"
            >
                <Input
                    inputValue={value}
                    onChange={handleValueChange}
                />
            </div>
        </div>
    )
}
