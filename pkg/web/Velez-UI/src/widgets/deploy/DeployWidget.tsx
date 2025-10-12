import cls from "@/widgets/deploy/DeployWidget.module.css"
import InfoMark from "@/assets/icons/InfoMark.svg";


import {Tooltip} from "react-tooltip";
import {ChangeEvent, useEffect, useState} from "react";
import JSONPretty from 'react-json-pretty';

import {CreateSmerdRequest} from "@vervstack/velez";

import Input from "@/components/base/Input.tsx";
import Checkbox from "@/components/base/Checkbox.tsx";

interface DeployWidgetProps {
    createSmerd: CreateSmerdRequest;
}

export default function DeployWidget({createSmerd}: DeployWidgetProps) {
    const [req, setReq] = useState<CreateSmerdRequest>({...createSmerd});

    useEffect(() => {
        createSmerd.name = createSmerd.name || ""
        createSmerd.imageName = createSmerd.imageName || ""

        createSmerd.ignoreConfig = createSmerd.ignoreConfig || false
        createSmerd.autoUpgrade = createSmerd.autoUpgrade || false
        createSmerd.useImagePorts = createSmerd.useImagePorts || false
        setReq(createSmerd)
    }, []);

    const updateField = (field: keyof CreateSmerdRequest, value: string | boolean) => {
        setReq(prev => ({
            ...prev,
            [field]: value
        }));
    };

    function stringFieldUpdater(field: keyof CreateSmerdRequest): (e: ChangeEvent<HTMLInputElement>) => void {
        return (e: ChangeEvent<HTMLInputElement>) => {
            updateField(field, e.target.value);
        }
    }

    function booleanFieldUpdater(field: keyof CreateSmerdRequest): (e: ChangeEvent<HTMLInputElement>) => void {
        return (e: ChangeEvent<HTMLInputElement>) => {
            console.log(e.target.checked)
            updateField(field, e.target.checked);
        }
    }

    return (
        <div className={cls.DeployWidgetContainer}>
            <div className={cls.ConfigurationInputs}>
                <Input
                    label="Name"
                    value={req.name || ''}
                    onChange={stringFieldUpdater("name")}
                />

                <Input
                    label="Image"
                    value={req.imageName || ''}
                    onChange={stringFieldUpdater("imageName")}
                />

                <Input
                    label="Command"
                    value={req.command || ''}
                    onChange={stringFieldUpdater("command")}
                />
                <div className={cls.CheckboxWrapper}>
                    <Checkbox
                        label="Ignore Matreshka Config"
                        onChange={booleanFieldUpdater("ignoreConfig")}
                        checked={req.ignoreConfig || false}/>

                    <div className={cls.InfoMarkTooltip}>
                        <img
                            src={InfoMark}
                            alt={'?'}
                            data-tooltip-id={"deploy-tooltip"}
                            data-tooltip-content="When deployed will be using default configuration from the inside of an image"
                        />
                    </div>
                </div>

                <div className={cls.CheckboxWrapper}>
                    <Checkbox
                        label="Use image ports"
                        onChange={booleanFieldUpdater("useImagePorts")}
                        checked={req.useImagePorts || false}/>

                    <div className={cls.InfoMarkTooltip}>
                        <img
                            src={InfoMark}
                            alt={'?'}
                            data-tooltip-id={"deploy-tooltip"}
                            data-tooltip-content="When checked - exposes all ports that image expose"
                        />
                    </div>
                </div>

                <div className={cls.CheckboxWrapper}>
                    <Checkbox
                        label="Auto upgrade"
                        onChange={booleanFieldUpdater("autoUpgrade")}
                        checked={req.autoUpgrade || false}/>

                    <div className={cls.InfoMarkTooltip}>
                        <img
                            src={InfoMark}
                            alt={'?'}
                            data-tooltip-id={"deploy-tooltip"}
                            data-tooltip-content="When checked automatically syncs new version of image and upgrades to it"
                        />
                    </div>
                </div>
            </div>

            <div className={cls.VervConfigBlock}>
                <JSONPretty data={req}/>
            </div>

            <Tooltip
                id={"deploy-tooltip"}
            />
        </div>
    );
}
