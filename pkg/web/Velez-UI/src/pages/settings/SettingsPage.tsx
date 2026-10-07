import cls from '@/pages/settings/SettingsPage.module.css';
import AddressesSettings from '@/widgets/settings/AddressesSettings/AddressesSettings.tsx';
import ApiSettings from '@/widgets/settings/ApiSettings/ApiSettings.tsx';
import EnvironmentsSettings from '@/widgets/settings/EnvironmentsSettings/EnvironmentsSettings.tsx';
import RegistriesSettings from '@/widgets/settings/RegistriesSettings/RegistriesSettings.tsx';
import SandboxSettings from '@/widgets/settings/SandboxSettings/SandboxSettings.tsx';

export default function SettingsPage() {
    return (
        <div className={cls.SettingsPageContainer}>
            <h1 className={cls.PageTitle}>Settings</h1>
            <div className={cls.SettingsGrid}>
                <ApiSettings/>
                <EnvironmentsSettings/>
                <RegistriesSettings/>
                <SandboxSettings/>
                <AddressesSettings/>
            </div>
        </div>
    );
}
