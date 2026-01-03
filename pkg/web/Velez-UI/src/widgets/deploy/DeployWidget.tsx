import cls from "@/widgets/deploy/DeployWidget.module.css"
import InfoMark from "@/components/base/InfoMark.tsx";

import {useState} from "react";
import ReactJsonView from '@microlink/react-json-view'

import Input from "@/components/base/Input.tsx";
import Checkbox from "@/components/base/Checkbox.tsx";
import Search from "@/components/base/Search.tsx";
import PlainMap from "@/components/base/PlainMap.tsx";
import PortsWidget from "@/widgets/PortsWidget.tsx";
import VolumesWidget from "@/widgets/VolumesWidget.tsx";

import {Bind, CreateSmerdReq, Port, Smerd, Volume} from "@/model/smerds/Smerds.ts";
import useSettings from "@/app/settings/state.ts";
import {DeploySmerd} from "@/processes/api/velez.ts";
import BindingWidget from "@/widgets/BindingsWidget.tsx";

interface DeployWidgetProps {
    createSmerdReq?: CreateSmerdReq;

    afterDeploy?: (req: CreateSmerdReq, smerd: Smerd) => void;
}

export default function DeployWidget({createSmerdReq, afterDeploy}: DeployWidgetProps) {
    const settings = useSettings();

    const [req, setReq] =
        useState<CreateSmerdReq>(createSmerdReq || new CreateSmerdReq());

    const updateField = (
        field: keyof CreateSmerdReq,
        value: string | boolean | null | Record<string, string> | Port[] | Volume[] | Bind[]) => {
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

    function booleanFieldUpdater(field: keyof CreateSmerdReq): (v: boolean) => void {
        return (v: boolean) => {
            updateField(field, v);
        }
    }

    function deploy() {
        DeploySmerd(req, settings.initReq())
            .then((smerd: Smerd) => {
                if (afterDeploy) {
                    afterDeploy(req, smerd)
                }
            })
    }

    return (
        <div className={cls.DeployWidgetContainer}>
            <div className={cls.InputAndDisplay}>
                <div className={cls.ConfigurationInputs}>
                    <div className={cls.InputWrapper}>
                        <Input
                            label="Name"
                            inputValue={req.name}
                            onChange={stringFieldUpdater("name")}
                        />
                    </div>
                    <div className={cls.InputWrapper}>
                        <Search
                            label="Image"
                            value={req.imageName}
                            onChange={stringFieldUpdater("imageName")}
                        />
                    </div>

                    <div className={cls.InputWrapper}>
                        <Input
                            label="Command"
                            inputValue={req.command}
                            onChange={stringFieldUpdater("command")}
                        />
                    </div>

                    <div className={cls.CheckboxWrapper}>
                        <Checkbox
                            label="Ignore Matreshka Config"
                            onChange={booleanFieldUpdater("ignoreConfig")}
                            checked={req.ignoreConfig || false}/>

                        <InfoMark
                            tooltip={"When deployed will be using default configuration from the inside of an image"}
                        />
                    </div>

                    <div className={cls.CheckboxWrapper}>
                        <Checkbox
                            label="Use image ports"
                            onChange={booleanFieldUpdater("useImagePorts")}
                            checked={req.useImagePorts || false}/>

                        <InfoMark
                            tooltip={"When checked - exposes all ports that image expose"}/>
                    </div>

                    <div className={cls.CheckboxWrapper}>
                        <Checkbox
                            label="Auto upgrade"
                            onChange={booleanFieldUpdater("autoUpgrade")}
                            checked={req.autoUpgrade || false}/>

                        <InfoMark
                            tooltip={"When checked automatically syncs new version of image and upgrades to it"}/>
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

                    <PortsWidget
                        ports={req.ports}
                        onChange={(p) => {
                            updateField('ports', p)
                        }}
                    />

                    <VolumesWidget
                        volumes={req.volumes}
                        onChange={(v) => {
                            updateField('volumes', v)
                        }}
                    />
                    <BindingWidget
                        bindings={req.binds}
                        onChange={(v) => {
                            updateField('binds', v)
                        }}
                    />
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
            </div>


            <div className={cls.Controls}>
                <button
                    onClick={deploy}
                    className={cls.DeployButton}>Deploy
                </button>
            </div>
        </div>
    );
}
