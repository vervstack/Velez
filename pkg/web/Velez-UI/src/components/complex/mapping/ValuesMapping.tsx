import cls from '@/components/complex/mapping/ValuesMapping.module.css';

import PlusButtonSVG from "@/assets/icons/Plus.svg";
import MinusButtonSVG from "@/assets/icons/Minus.svg";

import ActionButton from "@/components/base/ActionButton.tsx";
import {Children, ReactElement} from "react";

interface VolumesWidgetProps {
    children: ReactElement[];
    header: string;

    onAdd?: () => void;
    onDelete?: (index: number) => void;
}

export default function ValuesMapping({children, header, onAdd, onDelete}: VolumesWidgetProps) {
    return (
        <div className={cls.VolumesWidgetContainer}>
            <div className={cls.HeaderWrapper}>
                <div>{header}</div>
                {onAdd && <ActionButton iconPath={PlusButtonSVG} onClick={onAdd}/>}
            </div>

            <div className={cls.VolumeMappingWrapper}>
                {Children.map(children, (child, index) => (
                    <div className={cls.VolumePair} key={index}>
                        {child}

                        {onDelete &&
							<ActionButton iconPath={MinusButtonSVG}
							              onClick={() => onDelete(index)}/>}
                    </div>
                ))}

            </div>

        </div>
    )
}
