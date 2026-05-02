import cls from '@/components/base/SectionLabel.module.css';

interface SectionLabelProps {
    children: string;
}

export default function SectionLabel({ children }: SectionLabelProps) {
    return <div className={cls.SectionLabelContainer}>{children}</div>;
}
