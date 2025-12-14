import {useEffect, useState} from "react";
import cls from "@/widgets/vcn/RegisterVcnUserWidget.module.css";

import Input from "@/components/base/Input.tsx";
import Button from "@/components/base/Button.tsx";
import {ConnectUserToVcn} from "@/processes/api/verv_closed_network.ts";
import {useCredentialsStore} from "@/app/settings/creds.ts";

export default function RegisterVcnUserWidget() {
    const [key, setKey] = useState('')
    const [username, setUsername] = useState('')

    const [isValid, setIsValid] = useState(false)

    useEffect(() => {
        setIsValid(key.length > 0 && username.length > 0)
    }, [key, username])

    const credentialsStore = useCredentialsStore();

    function connectToVcn() {
        if (!isValid) throw 'key or username is empty'

        ConnectUserToVcn(credentialsStore.getInitReq(), key, username)
            .then()
    }

    return (
        <div className={cls.RegisterVcnUserWidgetContainer}>
            <div className={cls.Tip}>
                (?) Pass the key given when connecting to headscale
                via tailscale client and set the username for user
            </div>
            <Input
                label={'Key*'}
                inputValue={key}
                onChange={setKey}
            />

            <Input
                label={'Username*'}
                inputValue={username}
                onChange={setUsername}
            />

            <Button
                title={'Register'}
                isDisabled={!isValid}
                onClick={connectToVcn}
            />
        </div>
    )
}
