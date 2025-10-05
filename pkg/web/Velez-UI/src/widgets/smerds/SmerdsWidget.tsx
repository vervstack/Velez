import React, {useEffect, useState} from "react";

import {ListSmerdsRequest, Smerd} from "@vervstack/velez";

import useSettings from "@/app/settings/state.ts";
import {ListSmerds} from "@/processes/api/velez.ts";

import SmerdCard from "@/components/smerd/SmerdCard.tsx";
import Input from "@/components/base/Input.tsx";

export default function SmerdsWidget(): React.JSX.Element {
    const [smerds, setSmerds] =
        useState<Smerd[]>([])

    const [req, setReq] =
        useState<ListSmerdsRequest>({} as ListSmerdsRequest)

    const {initReq} = useSettings();

    useEffect(() => {
        ListSmerds(req, initReq())
            .then((resp) =>
                setSmerds(resp.smerds || []))
    }, [req]);
    return (
        <div>
            <SmerdsSearch req={req} setReq={setReq}/>
            <SmerdsList smerds={smerds}/>
        </div>
    )
}


function SmerdsSearch({req, setReq}: {
    req: ListSmerdsRequest,
    setReq: React.Dispatch<React.SetStateAction<ListSmerdsRequest>>
}): React.JSX.Element {

    return (<>
        <Input
            onChange={(elem) => {
                req.limit = Number(elem.target.value)
                setReq(req)
            }}
            value={0}
        />
    </>)
}

function SmerdsList({smerds}: { smerds: Smerd[] }): React.JSX.Element {
    return (<>
        {
            smerds.map((smerd, i) => {
                return (
                    <SmerdCard
                        key={i}
                        smerdInfo={smerd}/>
                )
            })}
    </>)

}
