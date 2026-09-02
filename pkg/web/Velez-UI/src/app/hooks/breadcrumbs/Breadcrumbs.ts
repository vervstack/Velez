import {create} from "zustand";

import type {Crumb} from "@/components/complex/BreadcrumbsBar/BreadcrumbsBar.tsx";

interface BreadcrumbsStore {
    crumbs: Crumb[];
    setCrumbs: (crumbs: Crumb[]) => void;
}

export const useBreadcrumbs = create<BreadcrumbsStore>((set) => ({
    crumbs: [],
    setCrumbs: (crumbs: Crumb[]) => set({crumbs}),
}));
