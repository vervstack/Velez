import {useState, useMemo} from "react"
import cn from "classnames"

import {useGetVervonomiconQuery} from "@/processes/queries/services"
import {buildVervonomiconTabs, findDefaultTab, shouldHighlightYaml} from "@/processes/vervonomicon"
import {useEnvironmentStore} from "@/app/hooks/environment/Environment"
import cls from "@/widgets/service/Vervonomicon/Vervonomicon.module.css"

interface Tab {
    id: string
    label: string
    filePath: string
    isResolved?: boolean
}

interface VervonomiconProps {
    serviceName: string
}

function highlightLines(src: string, isYaml: boolean): React.ReactNode[] {
    return src.split('\n').map(function renderLine(line: string, idx: number) {
        let content: React.ReactNode

        if (!isYaml) {
            content = line
        } else if (line.trimStart().startsWith('#')) {
            content = <span className={cls.SyntaxComment}>{line}</span>
        } else {
            const colonIdx = line.indexOf(':')
            if (colonIdx > 0) {
                const key = line.slice(0, colonIdx)
                const rest = line.slice(colonIdx)
                content = (
                    <>
                        <span className={cls.SyntaxKey}>{key}</span>
                        <span>{rest}</span>
                    </>
                )
            } else {
                content = line
            }
        }

        return (
            <div key={idx} className={cls.CodeLine}>
                <span className={cls.LineNumber}>{idx + 1}</span>
                <span className={cls.LineContent}>{content}</span>
            </div>
        )
    })
}

export default function Vervonomicon({serviceName}: VervonomiconProps) {
    const selectedEnvironment = useEnvironmentStore(state => state.selectedEnvironment)
    const {data: docs, isLoading} = useGetVervonomiconQuery(serviceName, selectedEnvironment)
    const [activeTabId, setActiveTabId] = useState<string>("")

    const tabs: Tab[] = useMemo(() => {
        if (!docs || docs.files.length === 0) return []
        const fileTabs = buildVervonomiconTabs(docs.files.map(f => f.path))
        const allTabs: Tab[] = [
            ...fileTabs,
            {id: "__resolved", label: "Resolved", filePath: "", isResolved: true},
        ]
        return allTabs
    }, [docs])

    const currentActiveTabId = useMemo(() => {
        if (activeTabId && tabs.some(t => t.id === activeTabId)) return activeTabId
        if (tabs.length === 0) return ""
        const defaultId = findDefaultTab(tabs.filter(t => !t.isResolved))
        return defaultId || (tabs[0]?.id ?? "")
    }, [activeTabId, tabs])

    const allEmpty = !isLoading && !!docs && docs.files.length === 0 && !docs.resolvedYaml

    function handleTabClick(tabId: string) {
        return function onClick() {
            setActiveTabId(tabId)
        }
    }

    function getActiveContent(): string {
        if (!docs) return ""
        if (currentActiveTabId === "__resolved") return docs.resolvedYaml
        const file = docs.files.find(f => f.path === currentActiveTabId)
        return file?.content ?? ""
    }

    function getActiveIsYaml(): boolean {
        if (currentActiveTabId === "__resolved") return true
        return shouldHighlightYaml(currentActiveTabId)
    }

    const activeContent = getActiveContent()
    const activeIsYaml = getActiveIsYaml()
    const lineCount = activeContent ? activeContent.split('\n').length : 0

    function getSourceLabel(source: string): string {
        if (source.includes("IMAGE")) return "image"
        if (source.includes("REPO")) return "repo"
        if (source.includes("PUSHED")) return "pushed"
        return "unknown"
    }

    return (
        <div className={cls.VervonomiconContainer}>
            <div className={cls.SectionHeader}>
                <h3 className={cls.SectionTitle}>Vervonomicon</h3>
                <span className={cls.SectionSubtitle}>
                    Declarative config and deployment manifests for this service
                </span>
            </div>

            {allEmpty ? (
                <p className={cls.Empty}>this service doesn't have a .verv/ descriptor</p>
            ) : (
                <div className={cls.Panel}>
                    <div className={cls.PanelHeader}>
                        {tabs.map(function renderTab(tab) {
                            return (
                                <button
                                    key={tab.id}
                                    className={cn(cls.TabBtn, currentActiveTabId === tab.id && cls.active)}
                                    onClick={handleTabClick(tab.id)}
                                    type="button"
                                >
                                    {tab.label}
                                </button>
                            )
                        })}
                        {docs && (docs.source || docs.environment) && (
                            <div className={cls.TabBadge}>
                                {docs.source && <span className={cls.BadgeLabel}>
                                    {getSourceLabel(docs.source)}
                                </span>}
                                {docs.environment && <span className={cls.BadgeLabel}>
                                    {docs.environment}
                                </span>}
                            </div>
                        )}
                    </div>

                    <div className={cls.CodeWrapper}>
                        {isLoading
                            ? <div className={cls.LoadingText}>Loading…</div>
                            : <pre className={cls.CodeBlock}>{highlightLines(activeContent, activeIsYaml)}</pre>
                        }
                    </div>

                    <div className={cls.FooterWrapper}>
                        <span className={cls.FooterMeta}>{lineCount} lines</span>
                        <span className={cls.FooterReadOnly}>read-only</span>
                    </div>
                </div>
            )}
        </div>
    )
}
