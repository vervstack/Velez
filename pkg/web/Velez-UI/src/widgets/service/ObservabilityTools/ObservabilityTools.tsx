import cls from '@/widgets/service/ObservabilityTools/ObservabilityTools.module.css'
import Button from "@/components/base/Button.tsx";

interface ObservabilityToolsProps {
    serviceName: string
}

interface Tool {
    id: string
    label: string
    sub: string
    icon: string
    accent: string
}

export default function ObservabilityTools({serviceName}: ObservabilityToolsProps) {
    const tools: Tool[] = [
        {id: 'logs', label: 'Logs', sub: 'Loki · last 1h', icon: '≈', accent: '#0ab7ee'},
        {id: 'grafana', label: 'Grafana', sub: `${serviceName} · prod`, icon: '◐', accent: '#f5a623'},
        {id: 'traces', label: 'Traces', sub: 'Tempo · p95 —ms', icon: '⥺', accent: '#9b7fd4'},
        {id: 'sentry', label: 'Sentry', sub: '—', icon: '◉', accent: '#ed2f32'},
        {id: 'git', label: 'Git', sub: '—', icon: '⎇', accent: '#b0b0c8'},
    ]

    return (
        <div className={cls.ObservabilityToolsContainer}>
            {tools.map(function renderTool(tool: Tool) {
                function handleClick() {
                    // TODO: navigate to tool
                }

                return (
                    // TODO: add observability support in Velez
                    <span
                        data-tooltip-id={'root-tooltip'}
                        data-tooltip-content={'Not available yet'}>
                        <Button
                            borderless
                            disabled
                        >
                            <div
                                key={tool.id}
                                className={cls.ToolCard}
                                style={{'--tool-accent': tool.accent} as React.CSSProperties}
                                onClick={handleClick}
                            >
                                <div className={cls.IconSquare}>{tool.icon}</div>
                                <div className={cls.TextWrapper}>
                                    <span className={cls.ToolLabel}>{tool.label}</span>
                                    <span className={cls.ToolSub}>{tool.sub}</span>
                                </div>
                                <span className={cls.ExternalArrow}>↗</span>
                            </div>
                        </Button>
                    </span>
                )
            })}
        </div>
    )
}
