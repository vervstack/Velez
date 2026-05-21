import {useState} from 'react';
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

    const [collapsed, setCollapsed] = useState(false);
    const [activeNodeId, setActiveNodeId] = useState<string | undefined>();
    const [showAllNodes, setShowAllNodes] = useState(false);


    const activeNav: NavId = ROUTE_TO_NAV[location.pathname] ?? 'apps';

    function handleCollapse() {
        setCollapsed(prev => !prev);
    }

    function handleNavChange(id: NavId) {
        navigate(NAV_TO_ROUTE[id]);
    }

    function handleToolNav(id: ToolId) {
        const route = TOOL_TO_ROUTE[id];
        if (route) {
            navigate(route);
        }
    }

    function handleToggleAllNodes() {
        setShowAllNodes(prev => !prev);
    }


    return (
        <div className={cls.MainLayoutContainer}>
            <Sidebar
                collapsed={collapsed}

                activeNodeId={activeNodeId}
                onNodeSelect={setActiveNodeId}

                activeNav={activeNav}
                onNavChange={handleNavChange}
                onToolNav={handleToolNav}
            />
            <div className={cls.ContentWithHeader}>
                <TopBar
                    collapsed={collapsed}
                    onCollapse={handleCollapse}
                    activeNodeId={activeNodeId}
                    showAllNodes={showAllNodes}
                    onToggleAllNodes={handleToggleAllNodes}
                    activeNav={activeNav}
                    onNavChange={handleNavChange}
                />

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
