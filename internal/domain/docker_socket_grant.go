package domain

// DockerSocketGrantSecretRef derives the Velez-internal secrets.Store key
// that gates create_smerd's host Docker socket bind-mount for one deploy -
// keyed by the deploy's exact service name, never by any client-supplied
// value, so no public API surface can influence it. Written only by trusted
// internal service code (runneraas.CreateRunner today) before scheduling a
// deploy; read by internal/workers/deploy_watcher.go immediately before it
// enqueues the resulting create_smerd task. See create_smerd.go's
// dockerSocketAccessor gate - this secret's mere presence is the grant, its
// value is never inspected.
func DockerSocketGrantSecretRef(serviceName string) SecretRef {
	return SecretRef{
		Scope: "plugin",
		Owner: "docker_socket_grant",
		Key:   serviceName,
	}
}
