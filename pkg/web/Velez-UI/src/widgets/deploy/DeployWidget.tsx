import cls from "@/widgets/deploy/DeployWidget.module.css"
import InfoMark from "@/assets/icons/InfoMark.svg";

import {Tooltip} from "react-tooltip";
import {useState} from "react";
import ReactJsonView from '@microlink/react-json-view'

import Input from "@/components/base/Input.tsx";
import Checkbox from "@/components/base/Checkbox.tsx";
import Search from "@/components/base/Search.tsx";
import PlainMap from "@/components/base/PlainMap.tsx";

import {ListImages} from "@/processes/api/velez.ts";
import useSettings from "@/app/settings/state.ts";
import {CreateSmerdReq} from "@/model/smerds/Smerds.ts";

interface DeployWidgetProps {
    createSmerdReq?: CreateSmerdReq;
}

export default function DeployWidget({createSmerdReq}: DeployWidgetProps) {
    const [req, setReq] =
        useState<CreateSmerdReq>(createSmerdReq || new CreateSmerdReq());

    const settings = useSettings()

    const updateField = (field: keyof CreateSmerdReq, value: string | boolean | null | Record<string, string>) => {
        setReq(prev => ({
            ...prev,
            [field]: value
        }));
    };

    function stringFieldUpdater(field: keyof CreateSmerdReq): (v: string) => void {
        return (v: string) => {
            if (v == '') {
                updateField(field, null);
                return
            }
            updateField(field, v);
        }
    }

    function booleanFieldUpdater(field: keyof CreateSmerdReq): (v: string) => void {
        return (v: string) => {
            updateField(field, v);
        }
    }

    function searchImage(v: string) {
        ListImages(v, settings.initReq())
            .then((r) => console.log(r))
        stringFieldUpdater("imageName")(v)
    }

    return (
        <div className={cls.DeployWidgetContainer}>
            <div className={cls.ConfigurationInputs}>
                <Input
                    label="Name"
                    inputValue={req.name}
                    onChange={stringFieldUpdater("name")}
                />
                <Search
                    label="Image"
                    value={req.imageName}
                    onChange={searchImage}
                />

                <Input
                    label="Command"
                    inputValue={req.command}
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

                <div>
                    <PlainMap
                        label={'Labels'}
                        records={req.labels || {}}
                        onChange={(newRecords) => {
                            updateField('labels', newRecords)
                        }}
                    />
                </div>

                <div>
                    <PlainMap
                        label={'Environment variables'}
                        records={req.env || {}}
                        onChange={(newRecords) => {
                            updateField('env', newRecords)
                        }}
                    />
                </div>
            </div>

            <div className={cls.VervConfigBlock}>
                <ReactJsonView
                    src={req}
                    theme={'flat'}
                    iconStyle={'triangle'}
                    name={null}
                    style={{
                        width: '100%',
                    }}
                    displayDataTypes={false}
                    displayObjectSize={false}
                    shouldCollapse={(field) => {
                        if (field.type == "array") {
                            // @ts-expect-error
                            return field.src.length === 0
                        }
                        return field.src == null;
                    }}
                />
            </div>

            <Tooltip
                id={"deploy-tooltip"}
            />
        </div>
    );
}
