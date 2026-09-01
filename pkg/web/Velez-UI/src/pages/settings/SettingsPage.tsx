import cls from '@/pages/settings/SettingsPage.module.css';
import ApiSettings from '@/widgets/settings/ApiSettings/ApiSettings.tsx';
import EnvironmentsSettings from '@/widgets/settings/EnvironmentsSettings/EnvironmentsSettings.tsx';
import RegistriesSettings from '@/widgets/settings/RegistriesSettings/RegistriesSettings.tsx';

export default function SettingsPage() {
    return (
        <div className={cls.SettingsPageContainer}>
            <h1 className={cls.PageTitle}>Settings</h1>
            <div className={cls.SettingsGrid}>
                <ApiSettings/>
                <EnvironmentsSettings/>
                <RegistriesSettings/>
            </div>
        </div>
    );
}
