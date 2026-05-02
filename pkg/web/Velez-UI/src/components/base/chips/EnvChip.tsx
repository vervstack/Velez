import cls from '@/components/base/chips/chips.module.css';
import cn from 'classnames';

interface EnvChipProps {
    env: string;
}

export default function EnvChip({ env }: EnvChipProps) {
    const isProd = env === 'prod';
    return (
        <span className={cn(cls.ChipContainer, isProd ? cls.envProd : cls.envStage)}>
            <span className={cn(cls.dot, isProd ? cls.dotProd : cls.dotStage)} />
            {env}
        </span>
    );
}
