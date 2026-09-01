import cls from '@/widgets/settings/RegistriesSettings/RegistriesSettings.module.css';
import type {Registry} from '@/app/api/velez';
import {RegistryType} from '@/app/api/velez';
import {ListRegistriesQuery} from "@/processes/queries/control_plane.ts";
import Button from "@/components/base/Button.tsx";
import {useDialog} from "@/app/hooks/dialog/Dialog.tsx";
import RegistryCreateDialog from "@/dialogs/RegistryCreateDialog/RegistryCreateDialog.tsx";
import RegistryManageDialog from "@/dialogs/RegistryManageDialog/RegistryManageDialog.tsx";

export default function RegistriesSettings() {
    const registriesQuery = ListRegistriesQuery()
    const {OpenDialog, CloseDialog} = useDialog();

    function handleManage(registry: Registry) {
        OpenDialog(
            <RegistryManageDialog
                registry={registry}
                onCancel={CloseDialog}
                onSaved={CloseDialog}
                onDeleted={CloseDialog}
            />
        )
    }

    function handleAddRegistry() {
        OpenDialog(
            <RegistryCreateDialog
                onCancel={CloseDialog}
                onCreated={CloseDialog}
            />
        )
    }

    const registries = registriesQuery.data?.registries || []

    return (
        <div className={cls.RegistriesSettingsContainer}>
            <div className={cls.SectionTitle}>Registries</div>
            {!registriesQuery.isLoading && (
                <div className={cls.CardsWrapper}>
                    {registries.map((r) =>
                        <RegistryRow key={r.id} registry={r} handleManage={handleManage}/>)}
                    {registries.length === 0 && (
                        <div className={cls.EmptyState}>No registries configured</div>
                    )}
                </div>)}
            <Button onClick={handleAddRegistry}>
                + Add registry
            </Button>
        </div>
    );
}

function registryTypeLabel(type?: RegistryType): string {
    if (type === RegistryType.REGISTRY_TYPE_GENERIC_V2) {
        return "Private registry"
    }
    return "Docker Hub"
}

function RegistryRow({registry, handleManage}: { registry: Registry, handleManage: (r: Registry) => void }) {
    function onManage() {
        handleManage(registry);
    }

    return (
        <div className={cls.RegistryRowWrapper}>
            <div className={cls.RegistryInfoWrapper}>
                <span className={cls.RegistryName}>{registry.name}</span>
                <span className={cls.RegistryType}>{registryTypeLabel(registry.type)}</span>
                {registry.isDefault && <span className={cls.DefaultBadge}>Default</span>}
            </div>
            <Button onClick={onManage}>
                Manage
            </Button>
        </div>
    );
}
