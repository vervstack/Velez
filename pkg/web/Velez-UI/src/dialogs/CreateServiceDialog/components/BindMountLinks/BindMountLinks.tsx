import cls from "@/dialogs/CreateServiceDialog/components/BindMountLinks/BindMountLinks.module.css"
import Input from "@/components/base/Input.tsx"
import {ResolvedLink} from "@/dialogs/CreateServiceDialog/processes/bindMounts.ts"

const EXPLANATION = "Every bind mount has to be linked. Velez creates a named volume bound to the same host " +
    "directory and uses it instead of the bind mount. The data stays where it is - nothing is copied."

interface Props {
    links: ResolvedLink[]
    onVolumeNameChange(destination: string, volumeName: string): void
}

export default function BindMountLinks({links, onVolumeNameChange}: Props) {
    function renderLink(link: ResolvedLink) {
        function handleChange(volumeName: string) {
            onVolumeNameChange(link.destination, volumeName)
        }

        return (
            <div key={link.destination} className={cls.LinkRow}>
                <span className={cls.Path}>{link.source} → {link.destination}</span>
                <Input label="Volume name" inputValue={link.volumeName} onChange={handleChange}/>
            </div>
        )
    }

    return (
        <div className={cls.BindMountLinksContainer}>
            <span className={cls.Title}>Bind mounts</span>
            <span className={cls.Explanation}>{EXPLANATION}</span>
            {links.map(renderLink)}
        </div>
    )
}
