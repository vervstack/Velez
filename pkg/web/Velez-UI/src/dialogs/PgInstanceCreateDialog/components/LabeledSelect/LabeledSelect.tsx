import cls from "@/dialogs/PgInstanceCreateDialog/components/LabeledSelect/LabeledSelect.module.css"

interface Option {
    value: string
    label: string
}

interface Props {
    label: string
    value: string
    placeholder: string
    options: Option[]
    onChange: (value: string) => void
    disabled?: boolean
}

export default function LabeledSelect({label, value, placeholder, options, onChange, disabled}: Props) {
    function handleChange(e: React.ChangeEvent<HTMLSelectElement>) {
        onChange(e.target.value)
    }

    function renderOption(opt: Option) {
        return <option key={opt.value} value={opt.value}>{opt.label}</option>
    }

    return (
        <div className={cls.LabeledSelectContainer}>
            <label className={cls.Label}>{label}</label>
            <select className={cls.Select} value={value} onChange={handleChange} disabled={disabled}>
                <option value="">{placeholder}</option>
                {options.map(renderOption)}
            </select>
        </div>
    )
}
