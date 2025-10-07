import {useEffect, useState} from "react";
import {useParams} from "react-router-dom";

import cls from "@/pages/smerd/SmerdPage.module.css";

import {GetSmerd} from "@/processes/api/velez.ts";
import {Smerd} from "@/model/smerds/Smerds.ts";
import useSettings from "@/app/settings/state.ts";

import Loader from "@/components/Loader.tsx";
import PortsWidget from "@/widgets/ports/PortsWidget.tsx";
import VolumesWidget from "@/widgets/volume/VolumesWidget.tsx";

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
    }, []);


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
                {smerdInfo.name}
            </div>
            <div className={cls.Body}>
                <div className={cls.InfoBox}>
                    <PortsWidget ports={smerdInfo.ports}/>
                </div>
                <div className={cls.InfoBox}>
                    <VolumesWidget volumes={smerdInfo.volumes}/>
                </div>
            </div>
        </div>
    )
}
