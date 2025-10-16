export function onUserReturn(callback: (tp: 'tab' | 'window') => void) {
    let hidden = false;

    document.addEventListener("visibilitychange", () => {
        if (document.visibilityState === "hidden") {
            hidden = true;
        } else if (document.visibilityState === "visible" && hidden) {
            hidden = false;
            callback("tab");
        }
    });

    window.addEventListener("focus", () => {
        callback("window");
    });
}
