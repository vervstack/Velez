import {KeyboardEvent} from "react"
import cn from "classnames"

import cls from "@/pages/container/components/ContainerLogLine/ContainerLogLine.module.css"
import {LogLine} from "@/processes/mappings/containerLogs.ts"
import ContainerLogHighlightedText from
    "@/pages/container/components/ContainerLogHighlightedText/ContainerLogHighlightedText.tsx"

interface Props {
    line: LogLine
    isShowTimestamp: boolean
    query: string
    isSelected: boolean
    onSelect: (line: LogLine) => void
}

export default function ContainerLogLine({line, isShowTimestamp, query, isSelected, onSelect}: Props) {
    function handleClick() {
        if (window.getSelection()?.toString()) return
        onSelect(line)
    }

    function handleKeyDown(e: KeyboardEvent<HTMLDivElement>) {
        if (e.key !== "Enter" && e.key !== " ") return
        e.preventDefault()
        onSelect(line)
    }

    return (
        <div
            role="button"
            tabIndex={0}
            className={cn(cls.ContainerLogLineContainer, {[cls.Stderr]: line.isStderr, [cls.Selected]: isSelected})}
            onClick={handleClick}
            onKeyDown={handleKeyDown}
        >
            {isShowTimestamp && line.timestamp && <span className={cls.Timestamp}>{line.timestamp}</span>}
            <ContainerLogHighlightedText text={line.text} query={query}/>
        </div>
    )
}
