const INSTANCE_NAME_PATTERN = /^[a-z0-9][a-z0-9_-]{1,31}$/
const INSTANCE_NAME_ERROR =
    "Instance name must be 2-32 characters: lowercase letters, digits, - and _, starting with a letter or digit"

export function validateInstanceName(name: string): string | undefined {
    const trimmed = name.trim()
    if (trimmed === "") {
        return undefined
    }
    return INSTANCE_NAME_PATTERN.test(trimmed) ? undefined : INSTANCE_NAME_ERROR
}
