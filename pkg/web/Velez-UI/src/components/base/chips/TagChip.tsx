import cls from '@/components/base/chips/chips.module.css';
import cn from 'classnames';

interface TagChipProps {
    tagKey: string;
    value: string;
}

export default function TagChip({ tagKey, value }: TagChipProps) {
    return (
        <span className={cn(cls.ChipContainer, cls.tag)}>
            {tagKey}: {value}
        </span>
    );
}
