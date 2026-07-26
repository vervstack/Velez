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

import {CreateSmerdReq, Port, Smerd, Volume} from "@/model/smerds/Smerds.ts";
import useSettings from "@/app/settings/state.ts";
import {DeploySmerdStream, GetSmerd} from "@/processes/api/velez.ts";
import {useToaster} from "@/app/hooks/toaster/Toaster.ts";
import {TaskStatus, TaskStatusStatus} from "@/app/api/velez";

interface DeployWidgetProps {
    createSmerdReq?: CreateSmerdReq;

    afterDeploy?: (req: CreateSmerdReq, smerd: Smerd) => void;
}

export default function DeployWidget({createSmerdReq, afterDeploy}: DeployWidgetProps) {
    const settings = useSettings();
    const toaster = useToaster();

    const [req, setReq] =
        useState<CreateSmerdReq>(createSmerdReq || new CreateSmerdReq());

    const [streamStatus, setStreamStatus] = useState<TaskStatusStatus | null>(null);

    const updateField = (
        field: keyof CreateSmerdReq,
        value: string | boolean | null | Record<string, string> | Port[] | Volume[] ) => {
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
        setStreamStatus(TaskStatusStatus.PENDING);

        let finalStatus: TaskStatusStatus | undefined;
        let finalError: string | undefined;

        function onStatus(status: TaskStatus) {
            finalStatus = status.status;
            finalError = status.error;
            if (status.status) {
                setStreamStatus(status.status);
            }
        }

        DeploySmerdStream(req, settings.initReq(), onStatus)
            .then(() => {
                if (finalStatus === TaskStatusStatus.FAILED) {
                    throw new Error(finalError || "smerd creation failed");
                }
                return GetSmerd(req.name, settings.initReq());
            })
            .then((smerd: Smerd) => {
                if (afterDeploy) {
                    afterDeploy(req, smerd)
                }
            })
            .catch(toaster.catchGrpc)
            .finally(() => {
                setStreamStatus(null);
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
                {streamStatus &&
                    <span className={cls.StreamStatus}>{formatStreamStatus(streamStatus)}</span>
                }
                <button
                    onClick={deploy}
                    className={cls.DeployButton}>Deploy
                </button>
            </div>
        </div>
    );
}

function formatStreamStatus(status: TaskStatusStatus): string {
    switch (status) {
        case TaskStatusStatus.PENDING:
            return "Deploying: pending…";
        case TaskStatusStatus.RUNNING:
            return "Deploying: running…";
        case TaskStatusStatus.DONE:
            return "Deploying: done";
        case TaskStatusStatus.FAILED:
            return "Deploying: failed";
        default:
            return "Deploying…";
    }
}
