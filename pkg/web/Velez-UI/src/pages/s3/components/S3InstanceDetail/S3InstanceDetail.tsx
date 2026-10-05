import {useState} from "react"

import cls from "@/pages/s3/components/S3InstanceDetail/S3InstanceDetail.module.css"
import Button from "@/components/base/Button.tsx"
import S3BucketsTab from "@/pages/s3/components/S3BucketsTab/S3BucketsTab.tsx"
import S3CredentialsTab from "@/pages/s3/components/S3CredentialsTab/S3CredentialsTab.tsx"
import S3KeysTab from "@/pages/s3/components/S3KeysTab/S3KeysTab.tsx"

type Tab = "buckets" | "keys" | "credentials"

const TABS: Array<{id: Tab, label: string}> = [
    {id: "buckets", label: "Buckets"},
    {id: "keys", label: "Keys"},
    {id: "credentials", label: "Credentials"},
]

interface Props {
    instanceName: string
}

export default function S3InstanceDetail({instanceName}: Props) {
    const [tab, setTab] = useState<Tab>("buckets")

    function renderTabButton(item: {id: Tab, label: string}) {
        function handleClick() {
            setTab(item.id)
        }

        return (
            <Button key={item.id} sm variant={tab === item.id ? "primary" : "secondary"} onClick={handleClick}>
                {item.label}
            </Button>
        )
    }

    function renderTab() {
        if (tab === "keys") return <S3KeysTab key={instanceName} instanceName={instanceName}/>
        if (tab === "credentials") return <S3CredentialsTab key={instanceName} instanceName={instanceName}/>
        return <S3BucketsTab key={instanceName} instanceName={instanceName}/>
    }

    return (
        <div className={cls.S3InstanceDetailContainer}>
            <div className={cls.Header}>
                <h2 className={cls.Title}>{instanceName}</h2>
                <div className={cls.Tabs}>
                    {TABS.map(renderTabButton)}
                </div>
            </div>
            {renderTab()}
        </div>
    )
}
