import PageHeader from "@/segments/PageHeader.tsx";
import {Outlet} from "react-router-dom";
import Toaster from "@/segments/Toaster.tsx";

export default function MainLayout() {
    return (
        <div
            style={{
                display: "flex",
                flexDirection: "column-reverse",
            }}
        >
            <Outlet/>
            <PageHeader/>
            <Toaster/>
        </div>
    )
}
