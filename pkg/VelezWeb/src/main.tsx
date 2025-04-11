import {createRoot} from 'react-dom/client'
import {RouterProvider} from "react-router-dom";

import '@/index.module.css'
import '@gravity-ui/uikit/styles/fonts.css';
import '@gravity-ui/uikit/styles/styles.css';
import 'react-tooltip/dist/react-tooltip.css'

import {ThemeProvider} from '@gravity-ui/uikit';

import router from "@/app/router/Router";

createRoot(document.getElementById('root')!).render(
    <ThemeProvider theme="light">
        <link href="https://fonts.googleapis.com/icon?family=Comfortaa" rel="stylesheet"/>
        <RouterProvider router={router}/>
    </ThemeProvider>
)
