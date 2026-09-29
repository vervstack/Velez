import {Toggle} from "@vervstack/chures"

import cls from "@/dialogs/CreateServiceDialog/components/PortOptions/PortOptions.module.css"
import type {Port} from "@/app/api/velez"
import Button from "@/components/base/Button.tsx"
import PortRowInputs from "@/dialogs/CreateServiceDialog/components/PortRowInputs/PortRowInputs.tsx"
import {describePort, EMPTY_PORT_ROW, PortRow} from "@/dialogs/CreateServiceDialog/processes/portRows.ts"

const KEEP_LABEL = "Keep existing ports (stops the container first → downtime)"
const KEEP_HINT = "Reusing the exact host ports is only possible through this toggle: the container is " +
    "stopped before it is recreated, so it is down until the new one is up."
const LIST_HINT = "Only the ports listed here are published. An empty list means no published ports."
const INVALID_TEXT = "Both ports of a row are required and must be whole numbers from 1 to 65535."

interface Props {
    publishedPorts: Port[]
    isKeepingPorts: boolean
    rows: PortRow[]
    isRowsInvalid: boolean
    onKeepChange(isKeeping: boolean): void
    onRowsChange(rows: PortRow[]): void
}

export default function PortOptions(
    {publishedPorts, isKeepingPorts, rows, isRowsInvalid, onKeepChange, onRowsChange}: Props
) {
    function handleAdd() {
        onRowsChange([...rows, EMPTY_PORT_ROW])
    }

    function renderPublished(port: Port) {
        return <span key={describePort(port)} className={cls.Published}>{describePort(port)}</span>
    }

    function renderRow(row: PortRow, index: number) {
        function handleChange(next: PortRow) {
            onRowsChange(rows.map((current, i) => (i === index ? next : current)))
        }

        function handleRemove() {
            onRowsChange(rows.filter((_, i) => i !== index))
        }

        return <PortRowInputs key={index} row={row} onChange={handleChange} onRemove={handleRemove}/>
    }

    return (
        <div className={cls.PortOptionsContainer}>
            <span className={cls.Title}>Ports</span>
            <div className={cls.PublishedWrapper}>
                <span className={cls.Hint}>Currently published (host → container):</span>
                {publishedPorts.map(renderPublished)}
            </div>
            <Toggle checked={isKeepingPorts} onChange={onKeepChange} label={KEEP_LABEL}/>
            <span className={cls.Hint}>{KEEP_HINT}</span>
            {!isKeepingPorts && (
                <div className={cls.RowsWrapper}>
                    <span className={cls.Hint}>{LIST_HINT}</span>
                    {rows.map(renderRow)}
                    {isRowsInvalid && <span className={cls.Error}>{INVALID_TEXT}</span>}
                    <Button variant="secondary" sm onClick={handleAdd}>Add port</Button>
                </div>
            )}
        </div>
    )
}
