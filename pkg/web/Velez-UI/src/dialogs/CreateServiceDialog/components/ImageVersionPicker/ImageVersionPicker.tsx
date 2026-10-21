import {useEffect, useMemo} from "react"
import {Dropdown, DropdownOption} from "@vervstack/chures"

import cls from "@/dialogs/CreateServiceDialog/components/ImageVersionPicker/ImageVersionPicker.module.css"
import {useImageVersionsQuery} from "@/processes/queries/containers.ts"
import Input from "@/components/base/Input.tsx"
import SkeletonLoader from "@/components/base/SkeletonLoader.tsx"
import {
    currentTagOf,
    pickVersionTag,
    repositoryOf,
    sortVersionTags,
} from "@/dialogs/CreateServiceDialog/processes/imageVersion.ts"

interface Props {
    containerId: string
    image: string
    value?: string
    onChange(tag: string): void
}

export default function ImageVersionPicker({containerId, image, value, onChange}: Props) {
    const query = useImageVersionsQuery(containerId)
    const currentTag = currentTagOf(image)

    const tags = useMemo(() => sortVersionTags(query.data?.tags ?? []), [query.data])
    const options = useMemo<DropdownOption[]>(() => tags.map((tag) => ({id: tag, name: tag})), [tags])
    const defaultTag = pickVersionTag(tags, currentTag)
    const selectedTag = value ?? defaultTag
    const isResolved = tags.length > 0
    const isImageKnown = image.length > 0

    useEffect(() => {
        if (isResolved && value === undefined) onChange(defaultTag)
    }, [isResolved, value, defaultTag])

    function handleChange(ids: string[]) {
        if (ids[0]) onChange(ids[0])
    }

    if (!isImageKnown) {
        return (
            <div className={cls.ImageVersionPickerContainer}>
                <div className={cls.Row}>
                    <SkeletonLoader shape="block" width="100%" height="2.5rem"/>
                    <SkeletonLoader shape="block" width="100%" height="2.5rem"/>
                </div>
            </div>
        )
    }

    return (
        <div className={cls.ImageVersionPickerContainer}>
            <div className={cls.Row}>
                <Input label="Image" inputValue={repositoryOf(image)}/>
                {isResolved
                    ? <Dropdown
                        label="Version"
                        options={options}
                        value={[selectedTag]}
                        onChange={handleChange}
                        portal
                    />
                    : <Input label="Version" inputValue={currentTag}/>}
            </div>
            {isResolved && selectedTag !== currentTag && (
                <span className={cls.Hint}>Pinned to this version when onboarded. The running image is the same.</span>
            )}
        </div>
    )
}
