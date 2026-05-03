import {CreateServiceRequest} from "@/app/api/velez";
import {useNavigate} from "react-router-dom";

import InitServiceWidget from "@/widgets/init_service/InitServiceWidget.tsx";
import {CreateService} from "@/processes/api/verv_api.ts";
import {useCredentialsStore} from "@/app/settings/creds.ts";
import {Routes} from "@/app/router/Router.tsx";

export default function NewServicePage() {
    const credentialsStore = useCredentialsStore();
    const navigate = useNavigate();

    async function createService(req: CreateServiceRequest) {
        await CreateService(credentialsStore.getInitReq(), req);
        navigate(Routes.Service + "/" + req.name);
    }

    return (
        <div>
            <InitServiceWidget createCallback={createService}/>
        </div>
    )
}
