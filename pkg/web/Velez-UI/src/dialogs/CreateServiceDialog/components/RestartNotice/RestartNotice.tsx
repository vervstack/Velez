import cn from "classnames"

import cls from "@/dialogs/CreateServiceDialog/components/RestartNotice/RestartNotice.module.css"

const CLUSTER_TEXT = "The container stays running. Velez onboards it without restarting or modifying it."
const SINGLE_NODE_TEXT = "The container will be recreated under a new container id. The old container stays " +
    "(paused or stopped, see below) until you finish onboarding."

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
