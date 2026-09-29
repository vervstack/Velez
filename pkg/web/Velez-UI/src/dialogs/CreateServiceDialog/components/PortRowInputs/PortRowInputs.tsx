import cls from "@/dialogs/CreateServiceDialog/components/PortRowInputs/PortRowInputs.module.css"
import Button from "@/components/base/Button.tsx"
import Input from "@/components/base/Input.tsx"
import {PortRow} from "@/dialogs/CreateServiceDialog/processes/portRows.ts"

interface Props {
    row: PortRow
    onChange(row: PortRow): void
    onRemove(): void
}

export default function PortRowInputs({row, onChange, onRemove}: Props) {
    function handleHostPortChange(hostPort: string) {
        onChange({...row, hostPort})
    }

    function handleContainerPortChange(containerPort: string) {
        onChange({...row, containerPort})
    }

    return (
        <div className={cls.PortRowInputsContainer}>
            <Input label="Host port" inputValue={row.hostPort} onChange={handleHostPortChange}/>
            <Input label="Container port" inputValue={row.containerPort} onChange={handleContainerPortChange}/>
            <Button variant="ghost" sm onClick={onRemove}>Remove</Button>
        </div>
    )
}
