import cls from "@/dialogs/CreateServiceDialog/components/PreviewSectionView/PreviewSectionView.module.css"
import PreviewRow from "@/dialogs/CreateServiceDialog/components/PreviewRow/PreviewRow.tsx"
import type {PreviewSection} from "@/dialogs/CreateServiceDialog/processes/registerPreview.ts"

interface Props {
    section: PreviewSection
}

export default function PreviewSectionView({section}: Props) {
    return (
        <div className={cls.PreviewSectionViewContainer}>
            <span className={cls.Title}>{section.title}</span>
            <ul className={cls.Rows}>
                {section.values.map(function renderValue(value, index) {
                    return <PreviewRow key={`${value.key}-${index}`} value={value}/>
                })}
            </ul>
        </div>
    )
}
