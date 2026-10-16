import Button from "@/components/base/Button.tsx"
import IconButton from "@/components/base/IconButton.tsx"
import {useToaster} from "@/app/hooks/toaster/Toaster.ts"
import ContainerLogHighlightedText from
    "@/pages/container/components/ContainerLogHighlightedText/ContainerLogHighlightedText.tsx"
import cls from "@/pages/container/components/ContainerLogDetailPanel/ContainerLogDetailPanel.module.css"
import ContainerMetaRow from "@/pages/container/components/ContainerMetaRow/ContainerMetaRow.tsx"
import {LogLine} from "@/processes/mappings/containerLogs.ts"

interface Props {
    line: LogLine | null
    query: string
    onClose: () => void
}

export default function ContainerLogDetailPanel({line, query, onClose}: Props) {
    const toaster = useToaster()

    if (!line) return null

    function handleCopy() {
        if (!line) return
        navigator.clipboard.writeText(line.text)
            .then(function notifyCopied() {
                toaster.bake({title: "Copied", description: "Log line copied to clipboard", level: "Info"})
            })
            .catch(toaster.catchGrpc)
    }

    return (
        <div className={cls.ContainerLogDetailPanelContainer}>
            <div className={cls.Header}>
                <span className={cls.Title}>Log line</span>
                <IconButton label="✕" title="Close" onClick={onClose}/>
            </div>
            <ContainerMetaRow label="Time">{line.timestamp || "—"}</ContainerMetaRow>
            <ContainerMetaRow label="Stream">{line.isStderr ? "stderr" : "stdout"}</ContainerMetaRow>
            <div className={cls.TextBlock}>
                <ContainerLogHighlightedText text={line.text} query={query}/>
            </div>
            <div className={cls.Actions}>
                <Button sm onClick={handleCopy}>Copy</Button>
            </div>
        </div>
    )
}
