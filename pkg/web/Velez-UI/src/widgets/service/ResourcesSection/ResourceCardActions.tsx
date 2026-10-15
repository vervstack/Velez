import {useDialog} from "@/app/hooks/dialog/Dialog.tsx"
import Button from "@/components/base/Button.tsx"
import type {ServiceResource} from "@/model/service_page/ServicePageModel"
import {resourceWebUiUrl} from "@/processes/resourceWebUi"
import ResourceAddressesDialog from "@/dialogs/ResourceAddressesDialog/ResourceAddressesDialog.tsx"

import cls from "./ResourceCard.module.css"

interface Props {
    resource: ServiceResource
}

export default function ResourceCardActions({resource}: Props) {
    const {OpenDialog} = useDialog()
    const hasAddresses = resource.addresses.length > 0
    const webUiUrl = hasAddresses ? undefined : resourceWebUiUrl(resource, window.location)

    function handleOpenAddresses() {
        OpenDialog(<ResourceAddressesDialog resourceName={resource.name} addresses={resource.addresses}/>)
    }

    function handleOpenWebUi() {
        if (webUiUrl) {
            window.open(webUiUrl, "_blank", "noopener,noreferrer")
        }
    }

    function handleStop(e: React.SyntheticEvent) {
        e.stopPropagation()
    }

    if (!hasAddresses && !webUiUrl) {
        return null
    }

    return (
        <div
            className={cls.ActionsWrapper}
            onClick={handleStop}
            onAuxClick={handleStop}
            onKeyDown={handleStop}
        >
            {hasAddresses && <Button variant="ghost" sm onClick={handleOpenAddresses}>Addresses</Button>}
            {webUiUrl && <Button variant="ghost" sm onClick={handleOpenWebUi}>Web UI</Button>}
        </div>
    )
}
