import cn from "classnames"

import cls from "@/dialogs/CreateServiceDialog/components/PortMappingTable/PortMappingTable.module.css"
import Input from "@/components/base/Input.tsx"
import {
    describeRowOutcome,
    PortMappingRow,
    rowOutcomeOf,
} from "@/dialogs/CreateServiceDialog/processes/portMapping.ts"

const VOLUMES_TEXT = "This container has volumes: it is stopped (not paused) before the new one starts to avoid " +
    "data corruption, and kept (not deleted) until you finish onboarding."
const INVALID_TEXT = "Each host port must be a whole number from 1 to 65535, and no two rows may share one."

interface Props {
    rows: PortMappingRow[]
    hasVolumes: boolean
    isInvalid: boolean
    onRowsChange(rows: PortMappingRow[]): void
}

export default function PortMappingTable({rows, hasVolumes, isInvalid, onRowsChange}: Props) {
    function renderCells(row: PortMappingRow, index: number) {
        function handleNewHostChange(newHost: string) {
            onRowsChange(rows.map((current, i) => (i === index ? {...current, newHost} : current)))
        }

        const key = `${row.containerPort}/${row.protocol}`
        return [
            <span key={key + ":container"} className={cls.Cell}>{row.containerPort}</span>,
            <span key={key + ":now"} className={cls.Cell}>{row.currentHost}</span>,
            <Input key={key + ":new"} inputValue={row.newHost} onChange={handleNewHostChange}/>,
        ]
    }

    function renderNote(row: PortMappingRow) {
        const className = cn(cls.Note, rowOutcomeOf(row) === "kept" && cls.Danger)
        return <span key={`${row.containerPort}/${row.protocol}`} className={className}>{describeRowOutcome(row)}</span>
    }

    return (
        <div className={cls.PortMappingTableContainer}>
            <div className={cls.Grid}>
                <span className={cls.Title}>Container port</span>
                <span className={cls.Title}>Host now</span>
                <span className={cls.Title}>Host new</span>
                {rows.flatMap(renderCells)}
            </div>
            <div className={cls.NotesWrapper}>
                {rows.map(renderNote)}
                {hasVolumes && <span className={cn(cls.Note, cls.Danger)}>{VOLUMES_TEXT}</span>}
                {isInvalid && <span className={cn(cls.Note, cls.Danger)}>{INVALID_TEXT}</span>}
            </div>
        </div>
    )
}
