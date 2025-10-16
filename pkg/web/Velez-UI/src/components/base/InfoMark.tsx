import cls from "@/components/base/InfoMark.module.css";
import InfoMarkSVG from "@/assets/icons/InfoMark.svg";

interface InfoMarkProps {
    tooltip?: string
}

export default function InfoMark({tooltip}: InfoMarkProps) {
    return (
        <div className={cls.InfoMarkTooltip}>
            <img
                src={InfoMarkSVG}
                alt={'?'}
                data-tooltip-id={"tooltip"}
                data-tooltip-content={tooltip}
            />
        </div>
    )
}
