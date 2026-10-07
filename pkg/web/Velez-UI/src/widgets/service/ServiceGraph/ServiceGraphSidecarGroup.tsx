import cls from "@/widgets/service/ServiceGraph/ServiceGraphSidecarGroup.module.css"
import type {ServiceSidecarView} from "@/model/service_page/ServicePageModel"
import {mapContainerStatusToDot} from "@/processes/mappings/containers.ts"
import {
    layoutSidecarGroup,
    sidecarDotTone,
    SIDECAR_LABEL_MAX_CHARS,
    truncateSidecarLabel,
} from "@/processes/mappings/serviceGraphLayout.ts"
import type {SidecarDotTone} from "@/processes/mappings/serviceGraphLayout.ts"

interface Props {
    sidecars: ServiceSidecarView[]
    centerX: number
    centerY: number
    ringRadius: number
    labelBottomY: number
}

const DOT_CLASSES: Record<SidecarDotTone, string> = {
    ok: cls.DotOk,
    warn: cls.DotWarn,
    bad: cls.DotBad,
    idle: cls.DotIdle,
}

const CHIP_RADIUS = 7
const DOT_RADIUS = 3.5
const DOT_INSET = 12
const LABEL_INSET = 22

export default function ServiceGraphSidecarGroup({sidecars, centerX, centerY, ringRadius, labelBottomY}: Props) {
    const layout = layoutSidecarGroup({
        centerX,
        centerY,
        ringRadius,
        labelBottomY,
        sidecarCount: sidecars.length,
    })
    const {outline} = layout

    return (
        <g className={cls.ServiceGraphSidecarGroupContainer}>
            <rect
                x={outline.x}
                y={outline.y}
                width={outline.width}
                height={outline.height}
                rx={14}
                className={cls.Outline}
            />
            <text x={layout.captionX} y={layout.captionY} className={cls.Caption}>service + sidecars</text>
            {layout.connectors.map(function renderConnector(d: string) {
                return <path key={d} d={d} className={cls.Connector}/>
            })}
            {layout.chips.map(function renderChip(chip) {
                const sidecar = sidecars[chip.index]
                const tone = sidecarDotTone(mapContainerStatusToDot(sidecar.status))
                const label = truncateSidecarLabel(sidecar.name)
                const isTruncated = sidecar.name.length > SIDECAR_LABEL_MAX_CHARS

                return (
                    <g key={sidecar.containerId}>
                        {isTruncated && <title>{sidecar.name}</title>}
                        <rect
                            x={chip.x}
                            y={chip.y}
                            width={chip.width}
                            height={chip.height}
                            rx={CHIP_RADIUS}
                            className={cls.Chip}
                        />
                        <circle
                            cx={chip.x + DOT_INSET}
                            cy={chip.y + chip.height / 2}
                            r={DOT_RADIUS}
                            className={DOT_CLASSES[tone]}
                        />
                        <text
                            x={chip.x + LABEL_INSET}
                            y={chip.y + chip.height / 2}
                            className={cls.ChipLabel}
                        >
                            {label}
                        </text>
                    </g>
                )
            })}
        </g>
    )
}
