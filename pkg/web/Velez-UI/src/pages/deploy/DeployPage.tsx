import cls from "@/pages/deploy/DeployPage.module.css";

import DeployWidget from "@/widgets/deploy/DeployWidget.tsx";
import {useLocation} from "react-router-dom";
import {CreateSmerdReq} from "@/model/smerds/Smerds.ts";

export default function DeployPage() {
    const location = useLocation();

    const req: CreateSmerdReq = location.state ? location.state.data as CreateSmerdReq : {} as CreateSmerdReq

    return (
        <div className={cls.DeployPageContainer}>
            <DeployWidget createSmerdReq={req}/>
        </div>
    )
}
