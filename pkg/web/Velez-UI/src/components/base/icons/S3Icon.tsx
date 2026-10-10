interface Props {
    className?: string
}

export default function S3Icon({className}: Props) {
    return (
        <svg
            className={className}
            aria-hidden="true"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            xmlns="http://www.w3.org/2000/svg"
        >
            <ellipse cx="12" cy="5" rx="9" ry="2.5"/>
            <path d="M3 5l2.2 13.2c.2 1.3 3.3 2.3 6.8 2.3s6.6-1 6.8-2.3L21 5"/>
            <path d="M4.2 12c.8 1 3.8 1.7 7.8 1.7s7-.7 7.8-1.7"/>
        </svg>
    )
}
