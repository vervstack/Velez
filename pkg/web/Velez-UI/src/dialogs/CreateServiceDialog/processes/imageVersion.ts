const DEFAULT_TAG = "latest"
const VERSION_TAG_PATTERN = /^v?\d+(\.\d+)*([-.][0-9A-Za-z]+)*$/
const NUMERIC_PREFIX_PATTERN = /^v?(\d+(?:\.\d+)*)/

function withoutDigest(image: string): string {
    const digestStart = image.indexOf("@")
    return digestStart === -1 ? image : image.slice(0, digestStart)
}

function tagSeparatorIndex(reference: string): number {
    const separator = reference.lastIndexOf(":")
    return separator > reference.lastIndexOf("/") ? separator : -1
}

export function repositoryOf(image: string): string {
    const reference = withoutDigest(image)
    const separator = tagSeparatorIndex(reference)
    return separator === -1 ? reference : reference.slice(0, separator)
}

export function currentTagOf(image: string): string {
    const reference = withoutDigest(image)
    const separator = tagSeparatorIndex(reference)
    if (separator === -1) return DEFAULT_TAG
    return reference.slice(separator + 1) || DEFAULT_TAG
}

export function isVersionTag(tag: string): boolean {
    return VERSION_TAG_PATTERN.test(tag)
}

function numericSegmentCount(tag: string): number {
    const match = NUMERIC_PREFIX_PATTERN.exec(tag)
    return match ? match[1].split(".").length : 0
}

function compareTags(a: string, b: string): number {
    const isAVersion = isVersionTag(a)
    const isBVersion = isVersionTag(b)
    if (isAVersion !== isBVersion) return isAVersion ? -1 : 1
    if (!isAVersion) return a.localeCompare(b)

    const bySegments = numericSegmentCount(b) - numericSegmentCount(a)
    if (bySegments !== 0) return bySegments
    const byLength = b.length - a.length
    return byLength !== 0 ? byLength : a.localeCompare(b)
}

export function sortVersionTags(tags: string[]): string[] {
    return [...tags].sort(compareTags)
}

export function pickVersionTag(tags: string[], current: string): string {
    const best = sortVersionTags(tags).find(isVersionTag)
    return best ?? current
}

export function imageTagToSend(selected: string, current: string): string | undefined {
    return selected === current ? undefined : selected
}
