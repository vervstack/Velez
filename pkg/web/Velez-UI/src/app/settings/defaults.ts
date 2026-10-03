export function defaultBackendUrl(): string {
    return import.meta.env.VITE_VELEZ_BACKEND_URL || window.location.origin
}
