import {useEffect, useState} from "react";
import {useParams} from "react-router-dom";

import cls from "@/pages/smerd/SmerdPage.module.css";

import {GetSmerd} from "@/processes/api/velez.ts";
import {Smerd} from "@/model/smerds/Smerds.ts";
import useSettings from "@/app/settings/state.ts";

import Loader from "@/components/Loader.tsx";

export default function SmerdPage() {
    const [smerdInfo, setSmerdInfo] = useState<Smerd>()

    const settings = useSettings();

    const params = useParams<Record<string, string>>();
    const smerdName = params['name']

    if (!smerdName) {
        throw 'No name provided'
    }

    useEffect(() => {
        GetSmerd(smerdName, settings.initReq())
            .then(setSmerdInfo)

    }, [settings]);


    if (!smerdInfo) {
        return (
            <div className={cls.SmerdPageContainer}>
                <Loader/>
            </div>
        )
    }

    return (
        <div className={cls.SmerdPageContainer}>
            <div className={cls.Header}>
                <div> {smerdInfo.name} </div>
            </div>
        </div>
    )
}
