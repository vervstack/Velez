import {create} from "zustand";
import React from "react";
import cls from "@/app/hooks/dialog/Dialog.module.css";

export interface DialogManager {
    children: React.JSX.Element[] | null;

    IsClickOffClosesDialog: boolean;

    LockClosing(): void;

    UnlockClosing(): void;

    OpenDialog(...children: React.JSX.Element[]): void;

    CloseDialog(): void;
}

export const useDialog =
    create<DialogManager>((set, get) => ({
        children: null,
        IsClickOffClosesDialog: true,

        LockClosing() {
            set({IsClickOffClosesDialog: false})
        },
        UnlockClosing() {
            set({IsClickOffClosesDialog: true})
        },

        OpenDialog(...children: React.JSX.Element[]) {
            set({children: children});
        },

        CloseDialog() {
            const {IsClickOffClosesDialog} = get()
            if (IsClickOffClosesDialog) {
                set({children: null})
            }
        }
    }))


export default function Dialog() {
    const {children, CloseDialog} = useDialog();

    if (!children) return null;

    return (
        <div
            className={cls.DialogContainer}
            onMouseDown={(e) => {
                if (e.target === e.currentTarget) CloseDialog();
            }}
        >
            {children.map((c, index) => {
                return (
                    <div key={index}>
                        {c}
                    </div>)
            })}
        </div>
    );
}

