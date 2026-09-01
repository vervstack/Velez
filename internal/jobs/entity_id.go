package jobs

// SmerdEntityID composes the jobs-engine entity id for a smerd create/upgrade
// task from the target environment's Docker suffix and the smerd's logical
// name.
//
// velez.tasks carries UNIQUE (entity_id, action). Keying a smerd task on the
// bare name alone makes two environments deploying the same service name
// collide on that key - the engine dedups them into one task and the second
// environment silently attaches to the first one's result. Folding the suffix
// in gives each environment its own row.
//
// An EMPTY suffix returns the name unchanged, with no separator: the
// single-environment / unconfigured-ContainerSuffix node - every caller that
// predates environments - keeps exactly the entity ids it has today, so
// existing velez.tasks rows and production keys are untouched.
func SmerdEntityID(suffix, name string) string {
	if suffix == "" {
		return name
	}

	return suffix + "/" + name
}
