import {VervPlugin} from "@/app/api/velez";


export default function UnknownPlugin({type}: VervPlugin) {
    return (
        <div>
            Unknown plugin type {type}
        </div>
    )
}
