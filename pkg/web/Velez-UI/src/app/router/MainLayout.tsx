import {useState} from 'react';
import {Outlet, useNavigate, useLocation} from 'react-router-dom';

import cls from '@/app/router/MainLayout.module.css';

import {NodeBaseInfo} from "@/app/api/velez";
import Sidebar from '@/widgets/sidebar/Sidebar';
import TopBar from '@/widgets/topbar/TopBar';
import Toaster from '@/segments/Toaster';
import {Routes} from '@/app/router/Routes';
import {ListNodes} from "@/processes/api/control_plane.ts";

import {useQuery} from "@tanstack/react-query";
import {useToaster} from "@/app/hooks/toaster/Toaster.ts";

type NavId = 'controlplane' | 'vcn' | 'deployments' | 'search';

const NAV_TO_ROUTE: Record<NavId, string> = {
    controlplane: Routes.ControlPlane,
    vcn: Routes.VCN,
    deployments: Routes.Deployments,
    search: Routes.Search,
};

const ROUTE_TO_NAV: Record<string, NavId> = {
    [Routes.ControlPlane]: 'controlplane',
    [Routes.VCN]: 'vcn',
    [Routes.Deployments]: 'deployments',
    [Routes.Search]: 'search',
};

export default function MainLayout() {
    const navigate = useNavigate();
    const location = useLocation();

    const [collapsed, setCollapsed] = useState(false);
    const [activeNodeId, setActiveNodeId] = useState<string | undefined>();
    const [showAllNodes, setShowAllNodes] = useState(false);

    const [nodes, setNodes] = useState<NodeBaseInfo[]>([]);

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

    const toaster = useToaster();

    const nodesQuery = useQuery({
        queryKey: ["nodes_main_layout"],
        queryFn: () => {
            ListNodes()
                .then((nodesList) => {
                    setNodes(nodesList.nodes || [])
                    return
                })
                .catch(toaster.catchGrpc)
        },
    })

    return (
        <div className={cls.MainLayoutContainer}>
            <Sidebar
                collapsed={collapsed}

                nodes={nodes}
                activeNodeId={activeNodeId}
                onNodeSelect={setActiveNodeId}
                isNodesLoading={nodesQuery.isLoading}

                activeNav={activeNav}
                onNavChange={handleNavChange}
            />
            <div className={cls.ContentWithHeader}>
                <TopBar
                    collapsed={collapsed}
                    onCollapse={handleCollapse}
                    nodes={nodes}
                    activeNodeId={activeNodeId}
                    showAllNodes={showAllNodes}
                    onToggleAllNodes={handleToggleAllNodes}
                    activeNav={activeNav}
                    onNavChange={handleNavChange}
                    onDeploy={handleDeploy}
                />

                <main className={cls.ContentWrapper}>
                    <Outlet/>
                </main>
            </div>

            <Toaster/>
        </div>
    );
}
