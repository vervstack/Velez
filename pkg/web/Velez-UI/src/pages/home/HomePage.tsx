import {useEffect, useState} from "react";

import {
    ServiceBaseInfo,
    ListServicesRequest,
    ListServicesResponse,

    Paging,
} from "@vervstack/velez"

import cls from '@/pages/home/Home.module.css';

import {ListServices} from "@/processes/api/service.ts";
import {useToaster} from "@/app/hooks/toaster/Toaster.ts";
import LoaderWithError from "@/components/LoaderWithError.tsx";
import {Routes} from "@/app/router/Router.tsx";
import {useNavigate} from "react-router-dom";

export default function HomePage() {
    const [isLoading, setIsLoading] = useState<boolean>(true);
    const [loadError, setLoadError] = useState<string | undefined>();

    const toaster = useToaster()
    const [list, setList] = useState<ServiceBaseInfo[]>([]);

    const [req] = useState<ListServicesRequest>({
        paging: {
            limit: '10',
            offset: '0',
        } as Paging,
        searchPattern: undefined,
    } as ListServicesRequest);

    function handleResponse(r: ListServicesResponse) {
        setList(r.services || [])
        // TODO handle paging and total
    }


    useEffect(() => {
        setIsLoading(false)
        ListServices(req)
            .then(handleResponse)
            .catch((e) => {
                setLoadError(e.toString())
                toaster.catchGrpc(e);
            })
            .finally(() => setIsLoading(false))
    }, [req, toaster]);


    if (isLoading || loadError != undefined) {
        return (<div className={cls.HomeContainer}><LoaderWithError err={loadError}/></div>)
    }

    return (
        <div className={cls.HomeContainer}>
            <div className={cls.ServicesDashboard}>
                {
                    list.length > 0 ?
                        <div>
                            {
                                list.map((service: ServiceBaseInfo) => {
                                    return (
                                        <ServiceCard
                                            {...service}
                                        />
                                    )
                                })
                            }
                        </div> :
                        <div>No services yet on this cluster</div>
                }
            </div>
        </div>
    )
}


function ServiceCard({name}: ServiceBaseInfo) {
    const navigate = useNavigate();

    function onClick() {
        navigate(Routes.Service + "/" + name)
    }

    return (
        <div
            className={cls.ServiceCardContainer}
            onClick={onClick}
        >
            {name}
        </div>
    )
}
