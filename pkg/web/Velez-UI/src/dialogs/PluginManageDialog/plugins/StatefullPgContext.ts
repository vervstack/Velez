export interface StatefullPgContext {
    exposePort: boolean;
    portNumber: string;
    isRunningInContainer: boolean;
}
