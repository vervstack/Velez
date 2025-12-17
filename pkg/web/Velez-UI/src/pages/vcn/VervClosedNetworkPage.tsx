import {useState} from "react";
import cn from "classnames";

import cls from '@/pages/vcn/VervClosedNetwork.module.css';

import RegisterVcnUserWidget from "@/widgets/vcn/RegisterVcnUserWidget.tsx";

enum DialogType {
    RegisterUser = 1
}

export default function VervClosedNetworkPage() {

    const [dialogType, setDialogType] = useState<DialogType | null>(null)

    return (
        <div className={cls.VcnPageContainer}>
            <div className={cls.QuickActions}>
                <div
                    className={cls.Action}
                     onClick={() => setDialogType(DialogType.RegisterUser)}
                >Connect new user</div>
            </div>

            <div className={cn(cls.Dialog, {
                [cls.opened]: dialogType !== null
            })}>
                <div className={cls.Content}>
                    <div>{
                        dialogType === DialogType.RegisterUser && <RegisterVcnUserWidget/>
                    }</div>
                </div>

                <div
                    className={cls.CloseButton}
                    onClick={() => setDialogType(null)}
                >X</div>
            </div>
        </div>
    )
}
