import {useToaster} from "@/app/hooks/toaster/Toaster.ts";
import {serviceService} from "@/processes/api/service.ts";
import Button from "@/components/base/Button.tsx";
import {VervPlugin} from "@/model/services/VervPlugins.tsx";
import cls from '@/dialogs/PluginManageDialog/PluginManageDialog.module.css';

interface StatefullPgDeadPromptProps {
    plugin: VervPlugin;
}

export default function StatefullPgDeadPrompt({plugin}: StatefullPgDeadPromptProps) {
    const toaster = useToaster();

    function handleRestart() {
        serviceService.restartService(plugin.serviceName)
            .catch(toaster.catchGrpc)
    }

    return (
        <div className={cls.ModalContainer}>
            <div className={cls.ModalContent}>
                <div className={cls.ActionSection}>
                    <span>Service is down. Run it?</span>
                    <Button onClick={handleRestart}>
                        Restart
                    </Button>
                </div>
            </div>
        </div>
    );
}
