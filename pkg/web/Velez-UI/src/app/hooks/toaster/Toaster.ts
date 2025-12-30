import {create} from "zustand";

export interface Toast {
    title: string;
    description: string

    level?: 'Error' | 'Warn' | 'Info'

    isDismissable?: boolean;
}

export interface Toaster {
    toasts: Toast[];

    bake: (t: Toast) => void;
    dismiss: (title: string) => void;

    catchGrpc: (e: Error) => void;
}

export const useToaster = create<Toaster>(
    (set, get) => ({
        toasts: [],

        bake: (newToast: Toast) => {

            const oldToast = get().toasts.find((t: Toast) => t.title === newToast.title)
            if (oldToast) {
                console.error(`Toast with title ${newToast.title} already exists`)
                return
            }
            set((state: Toaster) => ({toasts: [...state.toasts, newToast]}));

            setTimeout(() => {
                get().dismiss(newToast.title);
            }, 5000);
        },

        dismiss: (title: string) => {
            set((state) => ({
                toasts: state.toasts.filter((t: Toast) => t.title !== title),
            }));
        },
        catchGrpc: (e: Error) => {
            get().bake({title: "Error", description: e.message, level: 'Error'})
        }
    }));

