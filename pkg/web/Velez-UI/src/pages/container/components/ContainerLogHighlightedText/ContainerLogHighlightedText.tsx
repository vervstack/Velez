import {useMemo} from "react"

import cls from "@/pages/container/components/ContainerLogHighlightedText/ContainerLogHighlightedText.module.css"
import {QuerySegment, splitByQuery} from "@/processes/mappings/containerLogs.ts"

interface Props {
    text: string
    query: string
}

export default function ContainerLogHighlightedText({text, query}: Props) {
    const segments = useMemo(function splitText() {
        return splitByQuery(text, query)
    }, [text, query])

    function renderSegment(segment: QuerySegment, i: number) {
        if (!segment.isMatch) return segment.text
        return <mark key={i} className={cls.Match}>{segment.text}</mark>
    }

    return <span className={cls.ContainerLogHighlightedTextContainer}>{segments.map(renderSegment)}</span>
}
