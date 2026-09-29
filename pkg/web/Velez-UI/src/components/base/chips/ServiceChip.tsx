import cn from 'classnames';

import cls from '@/components/base/chips/chips.module.css';
import GitlabIcon from '@/components/base/icons/GitlabIcon';
import ServiceIcon from '@/components/base/icons/ServiceIcon';

const CHIP_ICONS = {
    gitlab: GitlabIcon,
    service: ServiceIcon,
} as const;

interface ServiceChipProps {
    serviceName: string;
    icon?: keyof typeof CHIP_ICONS;
    onClick: (serviceName: string) => void;
}

export default function ServiceChip({ serviceName, icon, onClick }: ServiceChipProps) {
    const Icon = icon ? CHIP_ICONS[icon] : null;

    function handleClick(e: React.MouseEvent) {
        e.stopPropagation();
        onClick(serviceName);
    }

    return (
        <button type="button" className={cn(cls.ChipContainer, cls.tag, cls.clickable)} onClick={handleClick}>
            {Icon && <Icon className={cls.icon} />}
            {serviceName}
        </button>
    );
}
