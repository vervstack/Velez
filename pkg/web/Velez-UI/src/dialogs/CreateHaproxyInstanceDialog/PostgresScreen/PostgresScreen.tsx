import {Button} from "@vervstack/chures"
import {useCallback, useState} from "react"

import type {PgBackend} from "@/app/api/velez/haproxy_api.pb"
import {PgBackendRole} from "@/processes/api/haproxy.ts"
import type {StepScreenProps} from "@/dialogs/StepsDialog/Step.ts"
import type {HaproxyWizardContext} from "@/dialogs/CreateHaproxyInstanceDialog/haproxyWizardContext.ts"
import cls from "@/dialogs/CreateHaproxyInstanceDialog/CreateHaproxyInstanceDialog.module.css"
import PgBackendEditor from "@/dialogs/CreateHaproxyInstanceDialog/PostgresScreen/PgBackendEditor/PgBackendEditor.tsx"
import PgBackendRow from "@/dialogs/CreateHaproxyInstanceDialog/PostgresScreen/PgBackendRow/PgBackendRow.tsx"

export default function PostgresScreen(props: StepScreenProps<HaproxyWizardContext>) {
    const {backends, sidecarContainerId, updateContext} = props
    const [isAdding, setIsAdding] = useState(false)
    const addBackend = useCallback((backend: PgBackend) => {
        updateContext({backends: [...backends, backend]})
        setIsAdding(false)
    }, [backends, updateContext])

    function setRole(index: number, role: PgBackendRole) {
        const updated = backends.map((backend, position) => position === index ? {...backend, role} : backend)
        updateContext({backends: updated})
    }

    function removeBackend(index: number) {
        const remaining = backends.filter((_, position) => position !== index)
        const updated = remaining.map((backend, position) =>
            position === 0 ? {...backend, role: PgBackendRole.PRIMARY} : backend)
        updateContext({backends: updated})
    }

    return (
        <div className={cls.Fields}>
            {backends.map((backend, index) => (
                <PgBackendRow key={backend.name} backend={backend} sidecarContainerId={sidecarContainerId}
                              isMaster={index === 0} onRoleChange={(role) => setRole(index, role)}
                              onRemove={() => removeBackend(index)}/>
            ))}
            {isAdding ? (
                <PgBackendEditor backends={backends} sidecarContainerId={sidecarContainerId}
                                 onAdd={addBackend} onCancel={() => setIsAdding(false)}/>
            ) : (
                <div className={cls.Actions}>
                    <Button variant="secondary" type="button" onClick={() => setIsAdding(true)}>
                        + Add instance
                    </Button>
                </div>
            )}
        </div>
    )
}
