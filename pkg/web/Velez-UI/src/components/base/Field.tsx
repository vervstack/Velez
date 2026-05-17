import cls from '@/components/base/Field.module.css';

interface FieldProps {
    label: string;
    required?: boolean;
    hint?: string;
    error?: string;
    children: React.ReactNode;
}

export default function Field({ label, required, hint, error, children }: FieldProps) {
    return (
        <div className={cls.FieldContainer}>
            <label className={cls.Label}>
                {label}
                {required && <span className={cls.Required}> *</span>}
            </label>
            {children}
            {error && <span className={cls.Error}>{error}</span>}
            {!error && hint && <span className={cls.Hint}>{hint}</span>}
        </div>
    );
}
