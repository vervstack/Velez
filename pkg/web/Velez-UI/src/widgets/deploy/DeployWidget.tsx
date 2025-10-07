import cls from "@/widgets/deploy/DeployWidget.module.css"

import { CreateSmerdRequest } from "@vervstack/velez";
import Input from "@/components/base/Input.tsx";
import {useState} from "react";

interface DeployWidgetProps {
    createSmerd: CreateSmerdRequest;
}

export default function DeployWidget({ createSmerd }: DeployWidgetProps) {
    const [req, setReq] = useState<CreateSmerdRequest>({ ...createSmerd });

    const updateField = (field: keyof CreateSmerdRequest, value: any) => {
        setReq(prev => ({
            ...prev,
            [field]: value
        }));
    };

    return (
        <div className={cls.DeployWidgetContainer}>
            <Input
                label="Name"
                value={req.name || ''}
                onChange={(e) => updateField("name", e.target.value)}
            />
        </div>
    );
}
