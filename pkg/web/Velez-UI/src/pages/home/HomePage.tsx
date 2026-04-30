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
import LoaderWrapper from "@/components/LoaderWrapper.tsx";
import {Routes} from "@/app/router/Router.tsx";
import {useNavigate} from "react-router-dom";
import Button from "@/components/base/Button.tsx";

export default function HomePage() {
    const [load, doLoad] = useState<Promise<void> | undefined>(undefined)

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
        doLoad(ListServices(req).then(handleResponse))

    }, [req, toaster]);

    function openCreateService() {

    }

    return (
        <div className={cls.HomeContainer}>
            <LoaderWrapper load={load}>
                {
                    list.length > 0 ?
                        <div className={cls.ServicesDashboard}>
                            <div>
                                {list.map((service: ServiceBaseInfo) => {
                                    return (<ServiceCard{...service}/>
                                    )
                                })
                                }
                            </div>
                        </div> :
                        <div className={cls.NoServicesMessageWrapper}>
                            <div>No services yet on this cluster. Want to deploy?</div>
                            <div className={cls.CreateServiceButton}>
                                <Button
                                    title={'Create a service'}
                                    onClick={openCreateService}
                                />
                            </div>
                        </div>
                }
            </LoaderWrapper>
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
