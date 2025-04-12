import {createBrowserRouter} from "react-router-dom";
import HomePage from "@/pages/home/HomePage";
import ErrorPage from "@/pages/error/ErrorPage";

const router = createBrowserRouter([
    {
        path: "/*",
        element: (<HomePage/>),
        errorElement: (<ErrorPage/>)
    },
]);


export default router
