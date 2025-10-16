import {createRoot} from 'react-dom/client'
import {RouterProvider} from "react-router-dom";

import '@/index.module.css'

import 'react-tooltip/dist/react-tooltip.css'

import SettingsWidget from "@/widgets/settings/SettingsWidget.tsx";

import router from "@/app/router/Router";
import {StrictMode} from "react";
import {Tooltip} from "react-tooltip";

createRoot(document.getElementById('root')!)
    .render(
        <StrictMode>
            <link href="@/assets/font/Comfortaa.ttf" rel="stylesheet"/>
            <RouterProvider router={router}/>
            <SettingsWidget/>
            <Tooltip
                id={"tooltip"}
            />
        </StrictMode>
    )
