import {useNavigate} from "react-router-dom";

import cls from "@/segments/PageHeader.module.css";

import VelezIcon from "@/assets/icons/services/velez.svg";
import {Routes} from "@/app/router/Router.tsx";

interface NavigationUnit {
    title: string,
    route: Routes
}

export default function PageHeader() {
    const navigate = useNavigate()

    const navigation: NavigationUnit[] = [
        {
            title: 'Control Plane',
            route: Routes.ControlPlane,
        },
        {
            title: 'Deploy',
            route: Routes.Deploy,
        },
    ]


    return (
        <div className={cls.PageHeaderContainer}>
            <div
                className={cls.HomeLogo}
                onClick={() => navigate(Routes.Home)}
            >
                <img src={VelezIcon} alt={'velez'}/>
            </div>

            <div className={cls.Navigation}>
                {navigation.map(u => {
                    return (
                        <div
                            key={u.title}
                            className={cls.NavElement}
                            onClick={() => navigate(u.route)}>
                            {u.title}
                        </div>
                    )
                })}
            </div>
        </div>
    )
}
