import PageHeader from "@/segments/PageHeader.tsx";
import {Outlet} from "react-router-dom";

export default function MainLayout(){
    return (
        <div>
            <PageHeader/>
            <Outlet/>
        </div>
    )
}
