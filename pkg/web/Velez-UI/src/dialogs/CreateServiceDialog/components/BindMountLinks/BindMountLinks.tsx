import {Checkbox} from "@vervstack/chures"

import cls from "@/dialogs/CreateServiceDialog/components/BindMountLinks/BindMountLinks.module.css"
import Input from "@/components/base/Input.tsx"
import {ResolvedLink} from "@/dialogs/CreateServiceDialog/processes/bindMounts.ts"

const EXPLANATION = "A bind mount is either linked to a named volume bound to the same host directory - " +
    "nothing is copied - or kept as a plain bind mount. System paths such as /dev and /var/run are kept as " +
    "bind mounts."

interface Props {
    links: ResolvedLink[]
    onVolumeNameChange(destination: string, volumeName: string): void
    onKeptChange(destination: string, isKept: boolean): void
}

export default function BindMountLinks({links, onVolumeNameChange, onKeptChange}: Props) {
    function renderLink(link: ResolvedLink) {
        function handleChange(volumeName: string) {
            onVolumeNameChange(link.destination, volumeName)
        }

        function handleKeptChange(isKept: boolean) {
            onKeptChange(link.destination, isKept)
        }

        return (
            <div key={link.destination} className={cls.LinkRow}>
                <div className={cls.Mapping}>
                    <span className={cls.Path}>{link.source}</span>
                    <span className={cls.Arrow} aria-hidden="true">→</span>
                    <span className={cls.Path}>{link.destination}</span>
                </div>
                <Checkbox
                    label="Keep as bind mount"
                    checked={link.isKeptAsBind}
                    onChange={handleKeptChange}
                />
                {!link.isKeptAsBind && (
                    <Input label="Volume name" inputValue={link.volumeName} onChange={handleChange}/>
                )}
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
