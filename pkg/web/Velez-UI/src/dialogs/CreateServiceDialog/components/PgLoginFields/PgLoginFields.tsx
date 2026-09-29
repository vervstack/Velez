import {Input} from "@vervstack/chures"

import cls from "@/dialogs/CreateServiceDialog/components/PgLoginFields/PgLoginFields.module.css"

const HINT = "The container has no POSTGRES_USER / POSTGRES_PASSWORD, so Velez needs an admin login. " +
    "It is tested before anything is changed."

interface Props {
    superuser: string
    password: string
    onSuperuserChange(superuser: string): void
    onPasswordChange(password: string): void
}

export default function PgLoginFields({superuser, password, onSuperuserChange, onPasswordChange}: Props) {
    return (
        <div className={cls.PgLoginFieldsContainer}>
            <span className={cls.Hint}>{HINT}</span>
            <Input label="Superuser" value={superuser} setValue={onSuperuserChange}/>
            <Input label="Password" type="password" value={password} setValue={onPasswordChange}/>
        </div>
    )
}
