import {useState} from "react"
import {useNavigate} from "react-router-dom"
import {useQueryClient} from "@tanstack/react-query"

import cls from "@/pages/login/LoginPage.module.css"
import {AUTH_GATE_QUERY_KEY} from "@/app/hooks/health/useAuthGate.ts"
import {useCredentialsStore} from "@/app/settings/creds.ts"
import Button from "@/components/base/Button.tsx"
import Input from "@/components/base/Input.tsx"

const DEFAULT_URL = import.meta.env.VITE_VELEZ_BACKEND_URL || "http://0.0.0.0:53891"

export default function LoginPage() {
    const navigate = useNavigate()
    const queryClient = useQueryClient()
    const credStore = useCredentialsStore()

    const [url, setUrl] = useState(credStore.url || DEFAULT_URL)
    const [token, setToken] = useState(credStore.token || "")

    function handleConnect() {
        credStore.setUrl(url)
        credStore.setToken(token)
        queryClient.resetQueries({queryKey: AUTH_GATE_QUERY_KEY})
        navigate("/", {replace: true})
    }

    return (
        <div className={cls.LoginPageContainer}>
            <div className={cls.LoginCardWrapper}>
                <h1 className={cls.Title}>Connect to Velez</h1>
                <p className={cls.Hint}>The node rejected the current credentials.</p>
                <Input label="Backend URL" inputValue={url} onChange={setUrl}/>
                <Input label="Auth header" inputValue={token} onChange={setToken}/>
                <Button variant="primary" fullWidth onClick={handleConnect}>Connect</Button>
            </div>
        </div>
    )
}
