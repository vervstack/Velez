import {useState} from "react";
import {CreateSmerdRequest} from "@/app/api/velez";

import cls from "@/widgets/deploy/DeployWidget.module.css"

import Input from "@/components/base/Input.tsx";


interface DeployWidgetProps {
    predefined?: CreateSmerdRequest;

    onCreate?: (req: CreateSmerdRequest) => void;
}

export default function DeploymentWidget({predefined, onCreate}: DeployWidgetProps) {
    const [req, setReq] =
        useState<CreateSmerdRequest>(predefined || {} as CreateSmerdRequest);
    const [validationError, setValidationError] = useState<string>("");

    function updateField(
        field: keyof CreateSmerdRequest,
        value: string | boolean | null) {
        setReq(prev => ({
            ...prev,
            [field]: value
        }));
        if (validationError) {
            setValidationError("");
        }
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

    function validateAndDeploy() {
        if (!req.imageName || req.imageName.trim() === "") {
            setValidationError("Image name is required");
            return;
        }

        if (onCreate) {
            onCreate(req);
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
                    {validationError && (
                        <div className={cls.ValidationError}>
                            {validationError}
                        </div>
                    )}
                </div>
            </div>

            <div className={cls.Controls}>
                <button
                    onClick={validateAndDeploy}
                    className={cls.DeployButton}>Deploy
                </button>
            </div>
        </div>
    );
}
