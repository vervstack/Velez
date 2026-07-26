import cls from '@/components/complex/QueryErrorState/QueryErrorState.module.css';
import Button from '@/components/base/Button.tsx';

interface QueryErrorStateProps {
    message?: string;
    onRetry: () => void;
}

export default function QueryErrorState({message, onRetry}: QueryErrorStateProps) {
    return (
        <div className={cls.QueryErrorStateContainer}>
            <div className={cls.Message}>{message ?? 'Failed to load data.'}</div>
            <Button variant="secondary" onClick={onRetry}>Retry</Button>
        </div>
    );
}
