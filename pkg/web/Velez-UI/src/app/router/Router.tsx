import {createBrowserRouter, Navigate} from "react-router-dom";
import HomePage from "@/pages/home/HomePage";
import ErrorPage from "@/pages/error/ErrorPage";
import ControlPlanePage from "@/pages/controlplane/ControlPlanePage.tsx";
import MainLayout from "@/app/router/MainLayout.tsx";
import AuthGate from "@/app/router/AuthGate.tsx";
import LoginPage from "@/pages/login/LoginPage.tsx";
import SmerdPage from "@/pages/smerd/SmerdPage.tsx";
import DeployPage from "@/pages/deploy/DeployPage.tsx";
import VervClosedNetworkPage from "@/pages/vcn/VervClosedNetworkPage.tsx";
import DeploymentsPage from "@/pages/deployments/DeploymentsPage";
import SearchPage from "@/pages/search/SearchPage";
import ServicesPage from "@/pages/services/ServicesPage";

import NewServicePage from "@/pages/service/NewServicePage.tsx";
import ServiceRouteDispatch from "@/pages/service/ServiceRouteDispatch.tsx";
import SettingsPage from "@/pages/settings/SettingsPage.tsx";
import PostgresPage from "@/pages/postgres/PostgresPage.tsx";

import {Routes, Arguments} from "@/app/router/Routes";

export {Routes, Arguments};

const router = createBrowserRouter([
    {
        path: Routes.Login,
        element: <LoginPage/>,
        errorElement: <ErrorPage/>,
    },
    {
        path: '/',
        element: <AuthGate/>,
        errorElement: <ErrorPage/>,
        children: [{
            element: <MainLayout/>,
            children: [
                {
                    index: true,
                    element: (<ServicesPage/>),
                },

            {
                path: Routes.Services,
                element: (<ServicesPage/>),
            },

            {
                path: "/apps",
                element: <Navigate to={Routes.Services} replace/>,
            },

            {
                path: Routes.Smerd,
                element: <HomePage/>,
            },

            {
                path: Routes.NewVervService,
                element: (<NewServicePage/>)
            },

            {
                path: Routes.Service + "/:" + Arguments.Key,
                element: (<ServiceRouteDispatch/>),
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
                path: Routes.Postgres,
                element: (<PostgresPage/>),
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
                path: Routes.Settings,
                element: (<SettingsPage/>),
            },

            {
                path: Routes.Smerd + "/:" + Arguments.Name,
                element: (<SmerdPage/>),
            },

            {
                path: '*',
                element: <Navigate to="/" replace/>,
            },
            ]
        }]
    },
]);

export default router
