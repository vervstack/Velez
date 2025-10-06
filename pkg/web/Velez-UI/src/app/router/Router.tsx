import {createBrowserRouter} from "react-router-dom";
import HomePage from "@/pages/home/HomePage";
import ErrorPage from "@/pages/error/ErrorPage";
import ControlPlanePage from "@/pages/controlplane/ControlPlanePage.tsx";
import MainLayout from "@/app/router/MainLayout.tsx";
import SmerdPage from "@/pages/smerd/SmerdPage.tsx";

export enum Routes {
    Home = "/",
    ControlPlane = "/cp",
    Deploy = "/deploy",
    Smerd = "/smerd"
}

export enum Arguments {
    Name = "name"
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
              path: Routes.Smerd+"/:"+Arguments.Name,
              element: (<SmerdPage/>),
            },
            {
                path: Routes.ControlPlane,
                element: (<ControlPlanePage/>),
            },
        ]
    },
]);

export default router
