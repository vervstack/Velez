import {Toggle} from "@vervstack/chures"

import Button from "@/components/base/Button.tsx"
import Input from "@/components/base/Input.tsx"
import cls from "@/pages/container/components/ContainerLogsControls/ContainerLogsControls.module.css"

export const TAIL_OPTIONS: number[] = [100, 500, 1000]

interface Props {
    tail: number
    onTailChange: (tail: number) => void
    isFollow: boolean
    onFollowChange: (isFollow: boolean) => void
    isShowTimestamps: boolean
    onShowTimestampsChange: (isShowTimestamps: boolean) => void
    filter: string
    onFilterChange: (filter: string) => void
}

export default function ContainerLogsControls(props: Props) {
    const {tail, onTailChange, isFollow, onFollowChange, filter, onFilterChange} = props

    function renderTailButton(option: number) {
        function handleClick() {
            onTailChange(option)
        }

        return (
            <Button key={option} sm variant={tail === option ? "primary" : "secondary"} onClick={handleClick}>
                {String(option)}
            </Button>
        )
    }

    return (
        <div className={cls.ContainerLogsControlsContainer}>
            <div className={cls.TailWrapper}>
                <span className={cls.TailLabel}>Tail</span>
                {TAIL_OPTIONS.map(renderTailButton)}
            </div>
            <Toggle label="Follow" labelPosition="right" checked={isFollow} onChange={onFollowChange}/>
            <Toggle
                label="Timestamps"
                labelPosition="right"
                checked={props.isShowTimestamps}
                onChange={props.onShowTimestampsChange}
            />
            <div className={cls.FilterWrapper}>
                <Input label="Filter" inputValue={filter} onChange={onFilterChange}/>
            </div>
        </div>
    )
}
