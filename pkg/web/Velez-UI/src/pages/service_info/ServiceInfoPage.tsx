import {useParams} from "react-router-dom";
import {useState} from "react";
import {useCredentialsStore} from "@/app/settings/creds.ts";
import {GetServiceById, GetServiceByName} from "@/processes/api/service.ts";
import {VervAppService} from "@vervstack/velez";


export default function ServiceInfoPage() {
    const params = useParams<Record<string, string>>();

    const [serviceInfo, setServiceInfo] = useState<VervAppService | null>(null)


    const credentialsStore = useCredentialsStore();

    function loadService() {
        const key = params['key'];
        if (!key) {
            // TODO redirect to new service
            throw 'No name or id provided'
        }

        const serviceId = Number(key)
        if (serviceId.toString() == key ) {
            GetServiceById(credentialsStore.getInitReq(), serviceId.toString())
                .then(setServiceInfo)
            return
        }

        GetServiceByName(credentialsStore.getInitReq(), key)
            .then(setServiceInfo)
    }

    loadService()
    return (
        <div>
            {serviceInfo ? (<div>{serviceInfo.name}</div>) : null}
        </div>
    )
}
