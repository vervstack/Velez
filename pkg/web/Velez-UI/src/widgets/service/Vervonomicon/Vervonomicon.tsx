import { useState } from 'react'
import { useGetVervonomiconQuery } from '@/processes/queries/services'
import type { VervonomiconDocs } from '@/model/service_page/ServicePageModel'
import cls from '@/widgets/service/Vervonomicon/Vervonomicon.module.css'

type TabKey = 'vervonomicon' | 'deployment' | 'configuration' | 'secrets'

const TABS: { key: TabKey; label: string }[] = [
    { key: 'vervonomicon',   label: 'vervonomicon.yaml'  },
    { key: 'deployment',     label: 'deployment.hcl'     },
    { key: 'configuration',  label: 'configuration.yaml' },
    { key: 'secrets',        label: 'secrets.yaml'       },
]

const STUB_PLACEHOLDER = '# No config available yet\n# TODO: implement GetVervonomicon RPC'

interface VervonomiconProps {
    serviceName: string
}

function highlightLines(src: string): React.ReactNode[] {
    return src.split('\n').map(function renderLine(line: string, idx: number) {
        let content: React.ReactNode

        if (line.trimStart().startsWith('#')) {
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

function isAllEmpty(docs: VervonomiconDocs): boolean {
    return !docs.vervonomicon && !docs.deployment && !docs.configuration && !docs.secrets
}

export default function Vervonomicon({serviceName}: VervonomiconProps) {
    const [activeTab, setActiveTab] = useState<TabKey>('vervonomicon')

    const {data: docs, isLoading} = useGetVervonomiconQuery(serviceName)

    const activeContent = docs
        ? (isAllEmpty(docs) ? STUB_PLACEHOLDER : (docs[activeTab] || STUB_PLACEHOLDER))
        : ''

    const lineCount = activeContent ? activeContent.split('\n').length : 0

    function handleTabClick(key: TabKey) {
        return function onClick() {
            setActiveTab(key)
        }
    }

    return (
        <div className={cls.VervonomiconContainer}>
            <div className={cls.HeaderWrapper}>
                <div className={cls.TitleGroup}>
                    <span className={cls.Title}>Vervonomicon</span>
                    <span className={cls.Subtitle}>{serviceName}</span>
                </div>

                <div className={cls.TabsWrapper}>
                    {TABS.map(function renderTab(tab) {
                        return (
                            <button
                                key={tab.key}
                                className={`${cls.TabBtn} ${activeTab === tab.key ? cls.active : ''}`}
                                onClick={handleTabClick(tab.key)}
                            >
                                {tab.label}
                            </button>
                        )
                    })}
                </div>
            </div>

            <div className={cls.CodeWrapper}>
                {isLoading
                    ? <div className={cls.LoadingText}>Loading…</div>
                    : <pre className={cls.CodeBlock}>{highlightLines(activeContent)}</pre>
                }
            </div>

            <div className={cls.FooterWrapper}>
                <span className={cls.FooterMeta}>{lineCount} lines</span>
                <span className={cls.FooterReadOnly}>read-only</span>
            </div>
        </div>
    )
}
