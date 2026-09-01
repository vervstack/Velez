import {Navigate, Outlet} from "react-router-dom"

import cls from "@/app/router/AuthGate.module.css"
import {useAuthGate} from "@/app/hooks/health/useAuthGate.ts"
import {Routes} from "@/app/router/Routes"

export default function AuthGate() {
    const status = useAuthGate()

    if (status === "unauthorized") {
        return <Navigate to={Routes.Login} replace/>
    }

    if (status === "checking") {
        return (
            <div className={cls.AuthGateContainer}>
                <div className={cls.Spinner}/>
            </div>
        )
    }

    return <Outlet/>
}
