import {Input} from "@vervstack/chures"

import cls from "@/dialogs/CreateServiceDialog/components/RegistryLoginFields/RegistryLoginFields.module.css"

const HINT = "The container has authentication configured, so Velez needs an existing login. " +
    "It is tested before anything is changed."

interface Props {
    username: string
    password: string
    onUsernameChange(username: string): void
    onPasswordChange(password: string): void
}

export default function RegistryLoginFields({username, password, onUsernameChange, onPasswordChange}: Props) {
    return (
        <div className={cls.RegistryLoginFieldsContainer}>
            <span className={cls.Hint}>{HINT}</span>
            <Input label="Username" value={username} setValue={onUsernameChange}/>
            <Input label="Password" type="password" value={password} setValue={onPasswordChange}/>
        </div>
    )
}
