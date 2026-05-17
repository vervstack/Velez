import cls from '@/components/base/Section.module.css';
import cn from 'classnames';

interface SectionProps {
    num?: string;
    label: string;
    sub?: string;
    children: React.ReactNode;
}

export default function Section({ num, label, sub, children }: SectionProps) {
    return (
        <section className={cls.SectionContainer}>
            <div className={cls.Header}>
                {num && <span className={cls.Num}>{num}</span>}
                <div className={cls.Titles}>
                    <span className={cls.Label}>{label}</span>
                    {sub && <span className={cls.Sub}>{sub}</span>}
                </div>
            </div>
            <div className={cn(cls.Body, { [cls.indented]: !!num })}>
                {children}
            </div>
        </section>
    );
}
