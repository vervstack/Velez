import cn from "classnames"

import cls from "@/pages/container/components/ContainerLogLine/ContainerLogLine.module.css"
import {LogLine} from "@/processes/mappings/containerLogs.ts"

interface Props {
    line: LogLine
    isShowTimestamp: boolean
}

export default function ContainerLogLine({line, isShowTimestamp}: Props) {
    return (
        <div className={cn(cls.ContainerLogLineContainer, {[cls.Stderr]: line.isStderr})}>
            {isShowTimestamp && line.timestamp && <span className={cls.Timestamp}>{line.timestamp}</span>}
            <span className={cls.Text}>{line.text}</span>
        </div>
    )
}
