import cls from "@/pages/service/widgets/ServiceComingSoon.module.css";

interface Props {
    label: string;
}

export default function ServiceComingSoon({label}: Props) {
    return (
        <div className={cls.ServiceComingSoonContainer}>
            <span className={cls.ComingSoonText}>{label} — coming soon</span>
        </div>
    );
}
