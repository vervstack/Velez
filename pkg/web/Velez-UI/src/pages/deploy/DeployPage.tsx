import cls from "@/pages/deploy/DeployPage.module.css";

import {CreateSmerdRequest} from "@vervstack/velez";

import DeployWidget from "@/widgets/deploy/DeployWidget.tsx";
import {useLocation} from "react-router-dom";

export default function DeployPage() {
    const location = useLocation();

    const req: CreateSmerdRequest = location.state.data || {}

    return (
        <div className={cls.DeployPageContainer}>
            <DeployWidget createSmerd={req}/>
        </div>
    )
}
