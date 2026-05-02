import {createRoot} from 'react-dom/client'
import {RouterProvider} from "react-router-dom";
import {Tooltip} from "react-tooltip";
import {QueryClient, QueryClientProvider} from "@tanstack/react-query";

import '@/index.module.css'
import 'react-tooltip/dist/react-tooltip.css'

import router from "@/app/router/Router";

const queryClient = new QueryClient();

createRoot(document.getElementById('root')!)
    .render(
        <QueryClientProvider client={queryClient}>
            <>
                <link href="@/assets/font/Comfortaa.ttf" rel="stylesheet"/>
                <RouterProvider router={router}/>
                <Tooltip id={"tooltip"}/>
            </>
        </QueryClientProvider>
    )
