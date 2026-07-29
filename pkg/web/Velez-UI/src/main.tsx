import {createRoot} from 'react-dom/client'
import {RouterProvider} from "react-router-dom";
import {QueryClientProvider} from "@tanstack/react-query";

import '@/index.css'
import 'react-tooltip/dist/react-tooltip.css'

import router from "@/app/router/Router";
import {queryClient} from "@/app/queryClient.ts";

createRoot(document.getElementById('root')!)
    .render(
        <QueryClientProvider client={queryClient}>
            <>
                <RouterProvider router={router}/>
            </>
        </QueryClientProvider>
    )
