import Button from "@/components/base/Button.tsx"
import type {ServiceResource} from "@/model/service_page/ServicePageModel"
import {resourceGoToUrl} from "@/processes/resourceGoTo"

import cls from "./ResourceGoToButton.module.css"

interface Props {
    resource: ServiceResource
}

export default function ResourceGoToButton({resource}: Props) {
    const url = resourceGoToUrl(resource.addresses, window.location)

    function handleOpen() {
        if (url) {
            window.open(url, "_blank", "noopener,noreferrer")
        }
    }

    function handleStop(e: React.SyntheticEvent) {
        e.stopPropagation()
    }

    if (!url) {
        return null
    }

    return (
        <div
            className={cls.ResourceGoToButtonContainer}
            onClick={handleStop}
            onAuxClick={handleStop}
            onKeyDown={handleStop}
        >
            <Button variant="ghost" sm onClick={handleOpen}>Go to</Button>
        </div>
    )
}
