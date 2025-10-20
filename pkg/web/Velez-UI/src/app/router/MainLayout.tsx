import PageHeader from "@/segments/PageHeader.tsx";
import {Outlet} from "react-router-dom";

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
        </div>
    )
}
