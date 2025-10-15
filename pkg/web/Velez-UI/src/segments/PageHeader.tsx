import {useNavigate} from "react-router-dom";

import cls from "@/segments/PageHeader.module.css";

import VelezIcon from "@/assets/icons/velez/VelezIcon.tsx";
import {Routes} from "@/app/router/Router.tsx";
import Input from "@/components/base/Input.tsx";

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

    function searchSmerd(v: string) {
        console.log(v)
    }


    return (
        <div className={cls.PageHeaderContainer}>
            <div
                className={cls.HomeLogo}
                onClick={() => navigate(Routes.Home)}
            >
                <VelezIcon/>
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

                <div className={cls.NavElement}>
                    <Input onChange={searchSmerd} inputValue={''}/>
                </div>
            </div>
        </div>
    )
}
