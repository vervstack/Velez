interface ServiceIconProps {
    className?: string;
}

export default function ServiceIcon({ className }: ServiceIconProps) {
    return (
        <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" xmlns="http://www.w3.org/2000/svg">
            <rect x="4" y="4" width="16" height="16" rx="3"/>
            <circle cx="12" cy="12" r="3"/>
        </svg>
    );
}
