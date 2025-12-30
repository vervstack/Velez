import {useState} from "react";
import cn from "classnames";

import cls from "@/widgets/services/EnableStatefullMode.module.css";

import {EnableStatefullCluster} from "@vervstack/velez";

import Button from "@/components/base/Button.tsx";
import Input from "@/components/base/Input.tsx";
import Toggle from "@/components/base/Toggle.tsx";

import DatabasePixel from "@/assets/icons/services/database-pixel.svg";

interface EnableStatefullModeProps {
    onDeploy: (r: EnableStatefullCluster) => Promise<void>
}

export default function EnableStatefullMode({onDeploy}: EnableStatefullModeProps) {
    const [isPortExposed, setIsPortExposed] = useState(false);
    const [exposeToPort, setExposeToPort] = useState<number | null>(null);

    const [isLoading, setIsLoading] = useState(false);

    function onPortInput(s: string) {
        if (s == "") {
            setExposeToPort(null)
        }

        if (!isStringANumber(s)) {
            return
        }
        const n = Number(s)
        if (n > 65535) {
            return;
        }

        setExposeToPort(Number(s))
    }

    return (
        <div className={cls.EnableStatefullModeContainer}>
            <div className={cls.HeaderWrapper}>
                <div className={cls.HeaderText}>Statefull mode</div>
                <img
                    className={cls.Logo}
                    src={DatabasePixel}
                    alt={'Statfull mode'}
                />
            </div>

            <div className={cls.SettingsWrapper}>
                <div
                    className={cls.PortToggleWrapper}
                    data-tooltip-id={"tooltip"}
                    data-tooltip-content={"Set to true if you want Postgres to be accessible via public internet"}
                    data-tooltip-place="top"
                >
                    <div>Expose Port</div>
                    <Toggle value={isPortExposed} onChange={setIsPortExposed}/>
                </div>

                <div className={cn(cls.ExposeToPortWrapper, {
                    [cls.enabled]: isPortExposed,
                })}>
                    <Input
                        disabled={!isPortExposed}
                        label={'Expose To Port'}
                        inputValue={exposeToPort ? exposeToPort.toString() : null}
                        onChange={onPortInput}
                        hint={"If not presented - exposed port will be assign randomly"}
                    />
                </div>

            </div>
            <Button
                isDisabled={isLoading}
                onClick={() => {
                    setIsLoading(true)
                    onDeploy({
                        isExposePort: isPortExposed,
                        exposeToPort: exposeToPort?.toString(),
                    } as EnableStatefullCluster)
                        .finally(() => setIsLoading(false))
                }}
                title={'Deploy'}/>
        </div>
    )
}


function isStringANumber(value: string): boolean {
    return !isNaN(Number(value)) && value.trim() !== ''
}
