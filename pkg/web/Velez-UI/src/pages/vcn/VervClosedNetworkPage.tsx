import {useSearchParams} from "react-router-dom"
import {Button} from "@vervstack/chures"

import {openCreateVcnNetworkDialog} from "@/dialogs/CreateVcnNetworkDialog/CreateVcnNetworkDialog.tsx"
import VcnNetworkDetails from "@/pages/vcn/components/VcnNetworkDetails/VcnNetworkDetails.tsx"
import VcnNetworkList from "@/pages/vcn/components/VcnNetworkList/VcnNetworkList.tsx"
import cls from "@/pages/vcn/VervClosedNetwork.module.css"

export default function VervClosedNetworkPage() {
    const [searchParams, setSearchParams] = useSearchParams()
    const networkId = searchParams.get("network")

    function handleBack() {
        const next = new URLSearchParams(searchParams)
        next.delete("network")
        setSearchParams(next)
    }

    function handleOpen(id: string) {
        const next = new URLSearchParams(searchParams)
        next.set("network", id)
        setSearchParams(next)
    }

    if (networkId) {
        return (
            <div className={cls.VervClosedNetworkPageContainer}>
                <div className={cls.Toolbar}>
                    <Button variant="secondary" onClick={handleBack}>Back to networks</Button>
                </div>
                <VcnNetworkDetails key={networkId} networkId={networkId}/>
            </div>
        )
    }

    return (
        <div className={cls.VervClosedNetworkPageContainer}>
            <header className={cls.Toolbar}>
                <h1 className={cls.PageTitle}>Verv Closed Networks</h1>
                <Button variant="primary" onClick={openCreateVcnNetworkDialog}>Create network</Button>
            </header>
            <VcnNetworkList onOpen={handleOpen}/>
        </div>
    )
}
