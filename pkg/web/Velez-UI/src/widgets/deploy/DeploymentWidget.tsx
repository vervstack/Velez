import {useState} from "react";
import {CreateSmerdRequest} from "@vervstack/velez";

import cls from "@/widgets/deploy/DeployWidget.module.css"

import Input from "@/components/base/Input.tsx";


interface DeployWidgetProps {
    predefined?: CreateSmerdRequest;

    onCreate?: (req: CreateSmerdRequest) => void;
}

export default function DeploymentWidget({predefined, onCreate}: DeployWidgetProps) {
    const [req, setReq] =
        useState<CreateSmerdRequest>(predefined || {} as CreateSmerdRequest);

    function updateField(
        field: keyof CreateSmerdRequest,
        value: string | boolean | null) {
        setReq(prev => ({
            ...prev,
            [field]: value
        }));
    }

    function stringFieldUpdater(field: keyof CreateSmerdRequest): (v: string) => void {
        return (v: string) => {
            if (v == '') {
                updateField(field, null);
                return
            }
            updateField(field, v);
        }
    }

    return (
        <div className={cls.DeployWidgetContainer}>
            <div className={cls.InputAndDisplay}>
                <div className={cls.ConfigurationInputs}>
                    <div className={cls.InputWrapper}>
                        <Input
                            label="Image"
                            inputValue={req.imageName || ""}
                            onChange={stringFieldUpdater("imageName")}
                        />
                    </div>
                </div>
            </div>

            <div className={cls.Controls}>
                {onCreate &&
					<button
						onClick={() => onCreate(req)}
						className={cls.DeployButton}>Deploy
					</button>}
            </div>
        </div>
    );
}
