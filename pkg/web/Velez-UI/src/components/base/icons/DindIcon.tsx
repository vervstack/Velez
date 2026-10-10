interface Props {
    className?: string
}

export default function DindIcon({className}: Props) {
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
            <rect x="3" y="3" width="18" height="18"/>
            <rect x="8" y="8" width="8" height="8"/>
            <path d="M3 3l5 5M21 3l-5 5M3 21l5-5M21 21l-5-5"/>
        </svg>
    )
}
