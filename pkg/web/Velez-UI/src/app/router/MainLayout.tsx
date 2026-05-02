import { useState } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import cls from '@/app/router/MainLayout.module.css';
import Sidebar from '@/widgets/sidebar/Sidebar';
import TopBar from '@/widgets/topbar/TopBar';
import Toaster from '@/segments/Toaster';
import { Routes } from '@/app/router/Router';

type NavId = 'controlplane' | 'vcn' | 'deployments' | 'search';

const MOCK_NODES = [
    { id: 'node01', host: '192.168.1.10', status: 'online'   as const },
    { id: 'node02', host: '192.168.1.42', status: 'online'   as const },
    { id: 'node03', host: '10.0.0.15',    status: 'degraded' as const },
];

const NAV_TO_ROUTE: Record<NavId, string> = {
    controlplane: Routes.ControlPlane,
    vcn:          Routes.VCN,
    deployments:  Routes.Deployments,
    search:       Routes.Search,
};

const ROUTE_TO_NAV: Record<string, NavId> = {
    [Routes.ControlPlane]: 'controlplane',
    [Routes.VCN]:          'vcn',
    [Routes.Deployments]:  'deployments',
    [Routes.Search]:       'search',
};

export default function MainLayout() {
    const navigate = useNavigate();
    const location = useLocation();

    const [collapsed, setCollapsed] = useState(false);
    const [activeNodeId, setActiveNodeId] = useState(MOCK_NODES[0].id);
    const [showAllNodes, setShowAllNodes] = useState(false);

    const activeNav: NavId = ROUTE_TO_NAV[location.pathname] ?? 'controlplane';

    function handleCollapse() {
        setCollapsed(prev => !prev);
    }

    function handleNavChange(id: NavId) {
        navigate(NAV_TO_ROUTE[id]);
    }

    function handleToggleAllNodes() {
        setShowAllNodes(prev => !prev);
    }

    function handleDeploy() {
        navigate(Routes.Deploy);
    }

    return (
        <div className={cls.MainLayoutContainer}>
            <Sidebar
                collapsed={collapsed}
                nodes={MOCK_NODES}
                activeNodeId={activeNodeId}
                onNodeSelect={setActiveNodeId}
                activeNav={activeNav}
                onNavChange={handleNavChange}
            />

            <div className={cls.content}>
                <TopBar
                    collapsed={collapsed}
                    onCollapse={handleCollapse}
                    nodes={MOCK_NODES}
                    activeNodeId={activeNodeId}
                    showAllNodes={showAllNodes}
                    onToggleAllNodes={handleToggleAllNodes}
                    activeNav={activeNav}
                    onNavChange={handleNavChange}
                    onDeploy={handleDeploy}
                />

                <main className={cls.main}>
                    <Outlet />
                </main>
            </div>

            <Toaster />
        </div>
    );
}
