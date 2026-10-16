import {KeyboardEvent, UIEvent, useEffect, useLayoutEffect, useMemo, useRef, useState} from "react"

import type {WatchContainerLogsResponse} from "@/app/api/velez"
import cls from "@/pages/container/components/ContainerLogsTab/ContainerLogsTab.module.css"
import {WatchContainerLogsStream} from "@/processes/api/velez.ts"
import {
    appendLogLines,
    filterLogLines,
    LogLine,
    MAX_LOG_LINES,
    toLogLine,
} from "@/processes/mappings/containerLogs.ts"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import QueryErrorState from "@/components/complex/QueryErrorState/QueryErrorState.tsx"
import ContainerLogDetailDrawer from
    "@/pages/container/components/ContainerLogDetailDrawer/ContainerLogDetailDrawer.tsx"
import ContainerLogLine from "@/pages/container/components/ContainerLogLine/ContainerLogLine.tsx"
import ContainerLogsControls from "@/pages/container/components/ContainerLogsControls/ContainerLogsControls.tsx"

type StreamStatus = "loading" | "live" | "ended" | "error"

const FLUSH_INTERVAL_MS = 100
const BOTTOM_THRESHOLD_PX = 24
const DEFAULT_TAIL = 500
const SKELETON_WIDTHS: string[] = ["90%", "70%", "80%", "55%", "85%", "65%", "75%", "50%"]

interface Props {
    containerId: string
}

export default function ContainerLogsTab({containerId}: Props) {
    const [tail, setTail] = useState(DEFAULT_TAIL)
    const [isFollow, setIsFollow] = useState(true)
    const [isShowTimestamps, setIsShowTimestamps] = useState(false)
    const [filter, setFilter] = useState("")
    const [lines, setLines] = useState<LogLine[]>([])
    const [status, setStatus] = useState<StreamStatus>("loading")
    const [attempt, setAttempt] = useState(0)
    const [selectedId, setSelectedId] = useState<number | null>(null)
    const [panelLine, setPanelLine] = useState<LogLine | null>(null)
    const [isPanelOpen, setIsPanelOpen] = useState(false)

    const viewportRef = useRef<HTMLDivElement>(null)
    const isPinnedRef = useRef(true)

    useEffect(function openLogStream() {
        const controller = new AbortController()
        let buffer: LogLine[] = []
        let timer = 0
        let nextId = 0

        setLines([])
        setStatus("loading")

        function flush() {
            window.clearTimeout(timer)
            timer = 0
            if (buffer.length === 0) return
            const batch = buffer
            buffer = []
            setLines(prev => appendLogLines(prev, batch, MAX_LOG_LINES))
            setStatus(prev => prev === "loading" ? "live" : prev)
        }

        function handleLine(response: WatchContainerLogsResponse) {
            buffer.push(toLogLine(response, nextId++))
            if (timer === 0) timer = window.setTimeout(flush, FLUSH_INTERVAL_MS)
        }

        WatchContainerLogsStream({id: containerId, tail, isFollow}, handleLine, controller.signal)
            .then(function handleClosed() {
                if (controller.signal.aborted) return
                flush()
                setStatus("ended")
            })
            .catch(function handleFailed() {
                if (controller.signal.aborted) return
                flush()
                setStatus("error")
            })

        return function closeLogStream() {
            controller.abort()
            window.clearTimeout(timer)
        }
    }, [containerId, tail, isFollow, attempt])

    const visibleLines = useMemo(function applyFilter() {
        return filterLogLines(lines, filter)
    }, [lines, filter])

    useLayoutEffect(function followBottom() {
        const viewport = viewportRef.current
        if (viewport && isPinnedRef.current) viewport.scrollTop = viewport.scrollHeight
    }, [visibleLines])

    const isViewportMounted = status !== "error"

    useEffect(function keepPinnedOnResize() {
        const viewport = viewportRef.current
        if (!viewport) return
        const observer = new ResizeObserver(function repin() {
            if (isPinnedRef.current) viewport.scrollTop = viewport.scrollHeight
        })
        observer.observe(viewport)
        return function stopObserving() {
            observer.disconnect()
        }
    }, [isViewportMounted])

    function handleScroll(e: UIEvent<HTMLDivElement>) {
        const el = e.currentTarget
        isPinnedRef.current = el.scrollHeight - el.scrollTop - el.clientHeight < BOTTOM_THRESHOLD_PX
    }

    function handleClosePanel() {
        setSelectedId(null)
        setIsPanelOpen(false)
    }

    function handleSelect(line: LogLine) {
        if (isPanelOpen && selectedId === line.id) {
            handleClosePanel()
            return
        }
        setSelectedId(line.id)
        setPanelLine(line)
        setIsPanelOpen(true)
    }

    function handleTailChange(value: number) {
        handleClosePanel()
        setTail(value)
    }

    function handleFollowChange(value: boolean) {
        handleClosePanel()
        setIsFollow(value)
    }

    function handleKeyDown(e: KeyboardEvent<HTMLDivElement>) {
        if (e.key === "Escape" && isPanelOpen) handleClosePanel()
    }

    function handleRetry() {
        isPinnedRef.current = true
        setAttempt(prev => prev + 1)
    }

    function renderSkeletonRow(width: string, i: number) {
        return (
            <div key={i} className={cls.SkeletonWrapper}>
                <SkeletonLoader shape="line" width={width} height="0.9rem"/>
            </div>
        )
    }

    function renderLine(line: LogLine) {
        return (
            <ContainerLogLine
                key={line.id}
                line={line}
                isShowTimestamp={isShowTimestamps}
                query={filter}
                isSelected={line.id === selectedId}
                onSelect={handleSelect}
            />
        )
    }

    function renderViewportContent() {
        if (status === "loading" && lines.length === 0) return SKELETON_WIDTHS.map(renderSkeletonRow)
        if (lines.length === 0) return <div className={cls.EmptyMessage}>No logs.</div>
        if (visibleLines.length === 0) return <div className={cls.EmptyMessage}>No matching lines.</div>
        return visibleLines.map(renderLine)
    }

    function renderViewport() {
        if (status === "error") {
            return <QueryErrorState message="Failed to stream container logs." onRetry={handleRetry}/>
        }
        return (
            <div ref={viewportRef} className={cls.Viewport} onScroll={handleScroll}>
                {renderViewportContent()}
            </div>
        )
    }

    return (
        <div className={cls.ContainerLogsTabContainer} onKeyDown={handleKeyDown}>
            <ContainerLogsControls
                tail={tail}
                onTailChange={handleTailChange}
                isFollow={isFollow}
                onFollowChange={handleFollowChange}
                isShowTimestamps={isShowTimestamps}
                onShowTimestampsChange={setIsShowTimestamps}
                filter={filter}
                onFilterChange={setFilter}
            />
            <div className={cls.Body}>
                {renderViewport()}
                <ContainerLogDetailDrawer
                    line={panelLine}
                    query={filter}
                    isOpen={isPanelOpen}
                    onClose={handleClosePanel}
                />
            </div>
        </div>
    )
}
