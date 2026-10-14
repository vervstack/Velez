import cls from "@/pages/auto-update/components/AutoUpdateDisabledBanner/AutoUpdateDisabledBanner.module.css"

export default function AutoUpdateDisabledBanner() {
    return (
        <div className={cls.AutoUpdateDisabledBannerContainer} role="alert">
            Auto-update worker is turned off by config (auto_update_is_enabled). Changes are saved on the
            services but nothing is updated until it is turned on.
        </div>
    )
}
