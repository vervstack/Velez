interface Props {
    className?: string
}

export default function OnboardedIcon({className}: Props) {
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
            <path d="M12 2.5l2.4 1.9 3-.2 1.1 2.8 2.6 1.6-.7 2.9 1.2 2.7-2.1 2.2-.4 3-3 .6-2.2 2.1L12 21l-2.5 1.1-2.2-2.1-3-.6-.4-3L1.8 14.2 3 11.5l-.7-2.9L4.9 7l1.1-2.8 3 .2z"/>
            <path d="M8.5 12.2l2.4 2.4 4.6-4.8"/>
        </svg>
    )
}
