# Configuration

# Local state

Local state - is a simple JSON file. 
By default, it is located at `/tmp/velez/local_state.json`
and this path can be configured via `local_state_path` matreshka variable
or `VELEZ_LOCAL_STATE_PATH` environment variable.

## VelezKey

Is an access Api key used for interaction.

Generated at the start of app and stored in local state file.
It must be passed via header `Grpc-Metadata-Authorization` for http call
or as grpc metadata as `Authorization` key.

[//]: # (TODO check if Authorization really works for grpc calls)

## MatreshkaKey

Is an access Api key used by Velez to make matreshka calls.
Stored mostly for internal usage.

## Network

Basic definition for networking supports multi-node network - Headscale if presented
and docker network by default for one node cluster.

### Headscale

Main multi-node cluster network. Consists of following fields

- `PgRootDsn` - dsn for cluster state postgres with root access (dml + ddl)
- `PgNodeDsn` - dsn for cluster state postgres with node level access (mostly dml)