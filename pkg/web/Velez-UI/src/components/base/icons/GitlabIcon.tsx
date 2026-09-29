interface GitlabIconProps {
    className?: string;
}

export default function GitlabIcon({ className }: GitlabIconProps) {
    return (
        <svg className={className} viewBox="0 0 24 24" fill="#FC6D26" xmlns="http://www.w3.org/2000/svg">
            <path d="M22.65 14.39L12 22.13 1.35 14.39a.84.84 0 0 1-.3-.94l1.22-3.78 2.44-7.51a.42.42 0 0 1 .81 0l2.44 7.49h8.1l2.44-7.51a.42.42 0 0 1 .81 0l2.44 7.51 1.22 3.78a.84.84 0 0 1-.32.96z"/>
        </svg>
    );
}
