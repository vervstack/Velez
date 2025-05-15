import {Smerd} from "@vervstack/velez";

export default function SmerdCard({smerdInfo}: { smerdInfo: Smerd }) {
    return (
        <div>
            <div>{smerdInfo.name}</div>
        </div>
    )
}
