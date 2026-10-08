export function taskKey(entityId: string, action: string): string {
    return `${entityId}/${action}`
}
