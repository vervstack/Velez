import {Outlet, useNavigate, useLocation} from 'react-router-dom';

import cls from '@/app/router/MainLayout.module.css';

import Sidebar from '@/widgets/sidebar/Sidebar';
import TopBar from '@/widgets/topbar/TopBar';
import Toaster from '@/segments/Toaster';
import {Routes} from '@/app/router/Routes';
import Dialog from "@/app/hooks/dialog/Dialog.tsx";
import {Tooltip} from "react-tooltip";

type NavId = 'controlplane' | 'vcn' | 'deployments' | 'apps' | 'search';
type ToolId = 'secrets' | 'config' | 'logs' | 'settings';

const NAV_TO_ROUTE: Record<NavId, string> = {
    controlplane: Routes.ControlPlane,
    vcn: Routes.VCN,
    deployments: Routes.Deployments,
    apps: Routes.Apps,
    search: Routes.Search,
};

const ROUTE_TO_NAV: Record<string, NavId> = {
    [Routes.ControlPlane]: 'controlplane',
    [Routes.VCN]: 'vcn',
    [Routes.Deployments]: 'deployments',
    [Routes.Apps]: 'apps',
    [Routes.Search]: 'search',
};

const TOOL_TO_ROUTE: Record<ToolId, string> = {
    settings: Routes.Settings,
    secrets: '',
    config: '',
    logs: '',
};

export default function MainLayout() {
    const navigate = useNavigate();
    const location = useLocation();

    const activeNav: NavId = ROUTE_TO_NAV[location.pathname] ?? 'apps';

    function handleNavChange(id: NavId) {
        navigate(NAV_TO_ROUTE[id]);
    }

    function handleToolNav(id: ToolId) {
        const route = TOOL_TO_ROUTE[id];
        if (route) {
            navigate(route);
        }
    }

    return (
        <div className={cls.MainLayoutContainer}>
            <Sidebar
                activeNav={activeNav}
                onNavChange={handleNavChange}
                onToolNav={handleToolNav}
            />
            <div className={cls.ContentWithHeader}>
                <TopBar/>

                <main className={cls.ContentWrapper}>
                    <Outlet/>
                </main>
            </div>

            <Dialog/>
            <Toaster/>
            <Tooltip id="root-tooltip"/>
        </div>
    );
}
