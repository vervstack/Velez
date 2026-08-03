import {create} from "zustand";
import type {Environment} from "@/app/api/velez";

// DEFAULT_ENVIRONMENT_NAME is the environment auto-selected on first load
// when the fetched list contains it. This preserves the experience of
// existing single-environment deployments, which only ever had "PROD".
export const DEFAULT_ENVIRONMENT_NAME = "PROD";

// EnvironmentStore is the global "currently selected cluster environment"
// (e.g. PROD / STAGING) used to scope requests across nodes/services.
//
// NOTE: this is unrelated to `ServiceEnvironment`
// (src/model/service_page/ServicePageModel.ts), which is a per-service
// dashboard status concept backing the local EnvSwitcher widget
// (src/widgets/service/EnvSwitcher/EnvSwitcher.tsx). Do not merge the two.
export interface EnvironmentStore {
    environments: Environment[];
    selectedEnvironment: string;

    setEnvironments: (environments: Environment[]) => void;
    selectEnvironment: (name: string) => void;
}

export const useEnvironmentStore = create<EnvironmentStore>((set, get) => ({
    environments: [],
    selectedEnvironment: "",

    setEnvironments: (environments: Environment[]) => {
        set({environments});

        const {selectedEnvironment} = get();
        const stillPresent = environments.some((e) => e.name === selectedEnvironment);
        if (selectedEnvironment && stillPresent) {
            return;
        }

        const preferred = environments.find((e) => e.name === DEFAULT_ENVIRONMENT_NAME);
        const fallback = preferred ?? environments[0];
        set({selectedEnvironment: fallback?.name ?? ""});
    },

    selectEnvironment: (name: string) => {
        set({selectedEnvironment: name});
    },
}));
