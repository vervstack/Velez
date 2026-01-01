import cls from "@/pages/service_info/parts/Header.module.css";
import Button from "@/components/base/Button.tsx";

interface HeaderProps {
    serviceName: string;
    onClickDeploy: () => void;
}

export default function Header({serviceName, onClickDeploy}: HeaderProps) {
    return (
        <div className={cls.ServiceNameHeaderContainer}>
            <div className={cls.ServiceNameHeader}> {serviceName}</div>
            <div className={cls.DeployButtonsWrapper}>
                <Button
                    title={'Deploy'}
                    onClick={onClickDeploy}
                />
            </div>
        </div>
    )
}
