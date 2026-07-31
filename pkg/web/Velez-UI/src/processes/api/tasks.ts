import {parseGrpcError} from "@vervstack/chures";

import {TasksApi, TaskStatus, WatchTaskRequest} from "@/app/api/velez";
import {useCredentialsStore} from "@/app/settings/creds.ts";

// WatchTaskStream wraps TasksApi.WatchTask: it opens a live stream of
// TaskStatus updates for the given job (identified by entityId + action)
// and forwards each update to onStatus as it arrives. The returned promise
// resolves once the stream ends (the caller is expected to have already
// observed the terminal status via onStatus) and rejects if the stream
// itself fails.
export async function WatchTaskStream(
    req: WatchTaskRequest,
    onStatus: (status: TaskStatus) => void
): Promise<void> {
    const initReq = useCredentialsStore.getState().getInitReq();

    try {
        return await TasksApi.WatchTask(req, onStatus, initReq);
    } catch (e) {
        throw parseGrpcError(e);
    }
}
