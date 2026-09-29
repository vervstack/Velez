import cn from "classnames"

import cls from "@/dialogs/CreateServiceDialog/components/RestartNotice/RestartNotice.module.css"

const CLUSTER_TEXT = "The container stays running. Velez registers it without restarting or modifying it."
const SINGLE_NODE_TEXT = "The container will be restarted. Velez recreates it under a new container id, " +
    "so it is unavailable while it starts again."

interface Props {
    isClusterMode: boolean
}

export default function RestartNotice({isClusterMode}: Props) {
    return (
        <div className={cn(cls.RestartNoticeContainer, isClusterMode ? cls.Info : cls.Danger)} role="note">
            {isClusterMode ? CLUSTER_TEXT : SINGLE_NODE_TEXT}
        </div>
    )
}
