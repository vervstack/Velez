import PageHeader from "@/segments/PageHeader.tsx";
import {Outlet} from "react-router-dom";
import Toaster from "@/segments/Toaster.tsx";
import ErrorBoundary from "@/components/complex/ErrorBoundary.tsx";

export default function MainLayout() {
    return (
        <div
            style={{
                display: "flex",
                flexDirection: "column-reverse",
            }}
        >
            <ErrorBoundary>
                <Outlet/>
            </ErrorBoundary>
            <PageHeader/>
            <Toaster/>
        </div>
    )
}
