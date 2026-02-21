package local_state

type ClusterState struct {
	PgRootDsn string `json:"PgRootDsn"`
	PgNodeDsn string `json:"PgNodeDsn"`
}
