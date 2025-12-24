import {createRoot} from 'react-dom/client'
import {RouterProvider} from "react-router-dom";
import {Tooltip} from "react-tooltip";

import '@/index.module.css'
import 'react-tooltip/dist/react-tooltip.css'

import SettingsWidget from "@/widgets/settings/SettingsWidget.tsx";

import router from "@/app/router/Router";

createRoot(document.getElementById('root')!)
    .render(
        <div>
            <link href="@/assets/font/Comfortaa.ttf" rel="stylesheet"/>
            <RouterProvider
                router={router}/>

            <SettingsWidget/>
            <Tooltip
                id={"tooltip"}
            />
        </div>
    )
