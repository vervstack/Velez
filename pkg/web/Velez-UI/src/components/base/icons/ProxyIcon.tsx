interface Props {
    className?: string
}

export default function ProxyIcon({className}: Props) {
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
            <circle cx="4" cy="12" r="1"/>
            <circle cx="20" cy="12" r="1"/>
            <path d="M8 8h9M14 5l3 3-3 3"/>
            <path d="M16 16H7M10 13l-3 3 3 3"/>
        </svg>
    )
}
