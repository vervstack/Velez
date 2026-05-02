import {createBrowserRouter} from "react-router-dom";
import HomePage from "@/pages/home/HomePage";
import ErrorPage from "@/pages/error/ErrorPage";
import ControlPlanePage from "@/pages/controlplane/ControlPlanePage.tsx";
import MainLayout from "@/app/router/MainLayout.tsx";
import SmerdPage from "@/pages/smerd/SmerdPage.tsx";
import DeployPage from "@/pages/deploy/DeployPage.tsx";
import VervClosedNetworkPage from "@/pages/vcn/VervClosedNetworkPage.tsx";
import DeploymentsPage from "@/pages/deployments/DeploymentsPage";
import SearchPage from "@/pages/search/SearchPage";

import NewServicePage from "@/pages/service/NewServicePage.tsx";
import ServiceInfoPage from "@/pages/service/ServiceInfoPage.tsx";

export enum Routes {
    Home = "/",
    ControlPlane = "/cp",
    Deploy = "/deploy",
    Deployments = "/deployments",
    Smerd = "/smerd",
    VCN = "/vcn",
    Search = "/search",
    NewVervService = '/new_verv_service',
    Service = '/service'
}

export enum Arguments {
    Name = "name",
    Key = "key"
}

const router = createBrowserRouter([
    {
        path: Routes.Home,
        element: <MainLayout/>,
        errorElement: <ErrorPage/>,
        children: [
            {
                index: true,
                element: <HomePage/>,
            },

            {
                path: Routes.NewVervService,
                element: (<NewServicePage/>)
            },

            {
                path: Routes.Service + "/:" + Arguments.Key,
                element: (<ServiceInfoPage/>),
            },
            {
                path: Routes.Deploy,
                element: (<DeployPage/>)
            },
            {
                path: Routes.ControlPlane,
                element: (<ControlPlanePage/>),
            },

            {
                path: Routes.Deployments,
                element: (<DeploymentsPage/>),
            },

            {
                path: Routes.VCN,
                element: (<VervClosedNetworkPage/>)
            },

            {
                path: Routes.Search,
                element: (<SearchPage/>),
            },

            {
                path: Routes.Smerd + "/:" + Arguments.Name,
                element: (<SmerdPage/>),
            },
        ]
    },
]);

export default router
