import {useEffect, useState} from "react"

import type {PgBackend} from "@/app/api/velez/haproxy_api.pb"
import {haproxyService} from "@/processes/api/haproxy.ts"

const VALIDATION_DELAY_MS = 500

export function usePgBackendValidation(
    backend: PgBackend | undefined,
    sidecarContainerId: string | undefined,
    onAvailable: (backend: PgBackend) => void,
) {
    const [attempt, setAttempt] = useState(0)
    const [isChecking, setIsChecking] = useState(false)
    const [error, setError] = useState("")

    useEffect(() => {
        setError("")
        setIsChecking(false)
        if (!backend) return

        let isCurrent = true
        const timer = window.setTimeout(async function validate() {
            setIsChecking(true)
            try {
                const result = await haproxyService.validatePgBackend({backend, sidecarContainerId})
                if (!isCurrent) return
                if (result.ok) onAvailable(backend)
                else setError(result.errorMessage || "PostgreSQL is unavailable.")
            } catch (failure) {
                if (isCurrent) setError(failure instanceof Error ? failure.message : String(failure))
            } finally {
                if (isCurrent) setIsChecking(false)
            }
        }, VALIDATION_DELAY_MS)

        return () => {
            isCurrent = false
            window.clearTimeout(timer)
        }
    }, [attempt, backend, onAvailable, sidecarContainerId])

    function retest() {
        setAttempt((value) => value + 1)
    }

    return {isChecking, error, retest}
}
