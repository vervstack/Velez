import {CreateServiceRequest} from "@vervstack/velez";

import InitServiceWidget from "@/widgets/init_service/InitServiceWidget.tsx";
import {CreateService} from "@/processes/api/verv_api.ts";
import {useCredentialsStore} from "@/app/settings/creds.ts";

export default function NewServicePage() {
    const credentialsStore = useCredentialsStore();

    async function createService(req: CreateServiceRequest) {
        await CreateService(credentialsStore.getInitReq(), req)
    }

    return (
        <div>
            <InitServiceWidget createCallback={createService}/>
        </div>
    )
}
