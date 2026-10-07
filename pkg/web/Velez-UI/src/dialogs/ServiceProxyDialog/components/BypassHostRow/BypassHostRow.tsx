import cls from "@/dialogs/ServiceProxyDialog/components/BypassHostRow/BypassHostRow.module.css"
import Button from "@/components/base/Button.tsx"
import Input from "@/components/base/Input.tsx"
import type {BypassHostRow as BypassHostRowModel} from "@/dialogs/ServiceProxyDialog/processes/buildProxyRequest.ts"

interface Props {
    row: BypassHostRowModel

    onChange(row: BypassHostRowModel): void

    onRemove(id: number): void
}

export default function BypassHostRow({row, onChange, onRemove}: Props) {
    function handleHostChange(host: string) {
        onChange({...row, host})
    }

    function handleRemove() {
        onRemove(row.id)
    }

    return (
        <div className={cls.BypassHostRowContainer}>
            <div className={cls.InputWrapper}>
                <Input label="Bypass host" inputValue={row.host} onChange={handleHostChange} placeholder="postgres"/>
            </div>
            <Button sm variant="ghost" onClick={handleRemove}>✕</Button>
        </div>
    )
}
