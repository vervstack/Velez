import type {VcnPeer} from "@/app/api/velez/verv_closed_network.pb"
import cls from "@/pages/vcn/components/VcnPeers/VcnPeers.module.css"

interface Props {
    peers?: VcnPeer[]
}

export default function VcnPeers({peers = []}: Props) {
    function renderPeer(peer: VcnPeer) {
        return (
            <tr key={peer.id}>
                <th scope="row">{peer.host || peer.id || "Unnamed peer"}</th>
                <td className={cls.Status}>{peer.status || "Unknown"}</td>
                <td>{peer.ipAddresses?.join(", ") || "—"}</td>
                <td className={cls.Type}>{peer.type || "—"}</td>
            </tr>
        )
    }

    function renderPeerTable() {
        if (peers.length === 0) return <p className={cls.EmptyMessage}>No peers have joined this network.</p>

        return (
            <table className={cls.Table} aria-label="Network peers">
                <thead>
                    <tr>
                        <th scope="col">Host</th>
                        <th scope="col">Status</th>
                        <th scope="col">IP addresses</th>
                        <th scope="col">Type</th>
                    </tr>
                </thead>
                <tbody>{peers.map(renderPeer)}</tbody>
            </table>
        )
    }

    return (
        <section className={cls.VcnPeersContainer}>
            <h2 className={cls.Title}>Peers ({peers.length})</h2>
            <div className={cls.TableWrapper}>{renderPeerTable()}</div>
        </section>
    )
}
