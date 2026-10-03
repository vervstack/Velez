import cls from "@/dialogs/CreateServiceDialog/components/NoPortsNotice/NoPortsNotice.module.css"
import {NO_PORTS_TEXT} from "@/dialogs/CreateServiceDialog/processes/restartMessage.ts"

export default function NoPortsNotice() {
    return <div className={cls.NoPortsNoticeContainer} role="note">{NO_PORTS_TEXT}</div>
}
