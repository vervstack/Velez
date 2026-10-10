interface Props {
    className?: string
}

export default function RegistryIcon({className}: Props) {
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
            <rect x="8" y="3" width="8" height="7" rx="1"/>
            <rect x="3" y="14" width="8" height="7" rx="1"/>
            <rect x="13" y="14" width="8" height="7" rx="1"/>
        </svg>
    )
}
