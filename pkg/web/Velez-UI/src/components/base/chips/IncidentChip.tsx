import cls from '@/components/base/chips/chips.module.css';

export default function IncidentChip() {
    return (
        <span className={cls.ChipContainer + ' ' + cls.incident}>
            <svg width="10" height="10" viewBox="0 0 10 10" fill="none" className={cls.icon}>
                <path
                    d="M5 1.2L9.2 8.8H0.8L5 1.2Z"
                    stroke="var(--red)"
                    strokeWidth="1.2"
                    strokeLinejoin="round"
                    fill="rgba(237,47,50,0.18)"
                />
                <line x1="5" y1="4.2" x2="5" y2="6.4" stroke="var(--red)" strokeWidth="1.1" strokeLinecap="round"/>
                <circle cx="5" cy="7.5" r="0.55" fill="var(--red)"/>
            </svg>
            incident
        </span>
    );
}
