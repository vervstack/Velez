import {useNavigate} from "react-router-dom";
import {useEffect, useRef, useState} from "react";

import cls from "@/segments/PageHeader.module.css";
import VelezIcon from "@/assets/icons/services/velez.svg";

import {Routes} from "@/app/router/Router.tsx";

import {ListSmerdsRequest} from "@vervstack/velez";

import {ListSmerds} from "@/processes/api/velez.ts";
import useSettings from "@/app/settings/state.ts";
import {SuggestElem} from "@/model/common/suggest.ts";

import InputSearch from "@/components/complex/search/InputSearch.tsx";

interface NavigationUnit {
    title: string,
    route: Routes
}

export default function PageHeader() {
    const navigate = useNavigate()

    const navigation: NavigationUnit[] = [
        {
            title: 'Control Plane',
            route: Routes.ControlPlane,
        },
        {
            title: 'Deploy',
            route: Routes.Deploy,
        },
    ]
    const settings = useSettings();

    const [search, setSearch] = useState('')
    const [suggestList, setSuggestList] = useState<SuggestElem[]>([])
    const searchRemoverRef = useRef<NodeJS.Timeout | null>(null);

    useEffect(() => {
        if (searchRemoverRef.current != null) return

        const req = {
            name: search,
        } as ListSmerdsRequest;

        ListSmerds(req, settings.initReq())
            .then((r) => {
                const suggests = (r.smerds || []).map((s) => {
                    return {
                        name: s.name,
                        link: Routes.Smerd + '/' + s.name
                    } as SuggestElem
                })

                setSuggestList(suggests)
            })
    }, [search]);


    function startSearchQueryRemover() {
        if (searchRemoverRef.current) clearInterval(searchRemoverRef.current);

        searchRemoverRef.current = setInterval(() => {
            setSearch(prev => {
                if (prev.length <= 0 && searchRemoverRef.current) {
                    clearInterval(searchRemoverRef.current);
                    searchRemoverRef.current = null;
                    return prev;
                }
                return prev.slice(0, -1);
            });
        }, 62);
    }

    return (
        <div className={cls.PageHeaderContainer}>
            <div
                className={cls.HomeLogo}
                onClick={() => navigate(Routes.Home)}
            >
                <img src={VelezIcon} alt={'velez'}/>
            </div>

            <div className={cls.Navigation}>
                {navigation.map(u => {
                    return (
                        <div
                            key={u.title}
                            className={cls.NavElement}
                            onClick={() => navigate(u.route)}>
                            {u.title}
                        </div>
                    )
                })}

                <div className={cls.NavElement}>
                    <div className={cls.ServiceSearch}>
                        <InputSearch
                            inputValue={search}
                            onChange={setSearch}
                            suggests={suggestList}
                            onSuggestDismiss={startSearchQueryRemover}
                        />
                    </div>
                </div>
            </div>

            <div className={cls.Settings}>Settings</div>
        </div>
    )
}
