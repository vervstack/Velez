import {useEffect, useState} from "react";
import {useParams} from "react-router-dom";

import cls from "@/pages/smerd/SmerdPage.module.css";

import {GetSmerd} from "@/processes/api/velez.ts";
import {Smerd} from "@/model/smerds/Smerds.ts";
import useSettings from "@/app/settings/state.ts";

import PortsWidget from "@/widgets/PortsWidget.tsx";
import VolumesWidget from "@/widgets/VolumesWidget.tsx";
import LoaderWrapper from "@/components/LoaderWrapper.tsx";

export default function SmerdPage() {
    const [smerdInfo, setSmerdInfo] = useState<Smerd>({} as Smerd)

    const [load, doLoad] = useState<Promise<void> | undefined>()

    const settings = useSettings();

    const params = useParams<Record<string, string>>();
    const smerdName = params['name']

    if (!smerdName) {
        throw 'No name provided'
    }

    useEffect(() => {
        doLoad(GetSmerd(smerdName, settings.initReq())
            .then(setSmerdInfo))
    }, []);


    return (
        <div className={cls.SmerdPageContainer}>
            <LoaderWrapper load={load}>
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
            </LoaderWrapper>
        </div>
    )
}
