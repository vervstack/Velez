import cls from '@/components/base/chips/chips.module.css';

export default function FreezeChip() {
    return (
        <span className={cls.ChipContainer + ' ' + cls.freeze}>
            <svg width="10" height="10" viewBox="0 0 10 10" fill="none" className={cls.icon}>
                <line x1="5" y1="1" x2="5" y2="9" stroke="var(--blue)" strokeWidth="1.2" strokeLinecap="round"/>
                <line x1="1" y1="5" x2="9" y2="5" stroke="var(--blue)" strokeWidth="1.2" strokeLinecap="round"/>
                <line x1="2.05" y1="2.05" x2="7.95" y2="7.95" stroke="var(--blue)" strokeWidth="1.2" strokeLinecap="round"/>
                <line x1="7.95" y1="2.05" x2="2.05" y2="7.95" stroke="var(--blue)" strokeWidth="1.2" strokeLinecap="round"/>
                <circle cx="5" cy="1" r="0.9" fill="var(--blue)"/>
                <circle cx="5" cy="9" r="0.9" fill="var(--blue)"/>
                <circle cx="1" cy="5" r="0.9" fill="var(--blue)"/>
                <circle cx="9" cy="5" r="0.9" fill="var(--blue)"/>
            </svg>
            freeze
        </span>
    );
}
