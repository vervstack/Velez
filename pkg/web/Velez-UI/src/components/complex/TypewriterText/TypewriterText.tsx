import {useEffect, useState} from "react"
import cn from "classnames"

import cls from "@/components/complex/TypewriterText/TypewriterText.module.css"
import {
    frameDelayMs,
    nextTypewriterFrame,
} from "@/components/complex/TypewriterText/processes/nextTypewriterFrame.ts"

interface Props {
    text: string
}

function readPrefersReducedMotion(): boolean {
    return window.matchMedia?.("(prefers-reduced-motion: reduce)").matches ?? false
}

export default function TypewriterText({text}: Props) {
    const [isReducedMotion] = useState(readPrefersReducedMotion)
    const [shown, setShown] = useState("")

    useEffect(() => {
        if (isReducedMotion || shown === text) {
            return
        }

        const next = nextTypewriterFrame(shown, text)
        const timer = setTimeout(() => setShown(next), frameDelayMs(shown, next))

        return () => clearTimeout(timer)
    }, [shown, text, isReducedMotion])

    const displayed = isReducedMotion ? text : shown
    const isAnimating = displayed !== text

    return (
        <span className={cn(cls.TypewriterTextContainer, {[cls.IsAnimating]: isAnimating})}>
            {displayed}
            <span className={cls.Caret}/>
        </span>
    )
}
