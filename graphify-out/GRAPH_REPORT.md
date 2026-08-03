# Graph Report - Velez  (2026-08-03)

## Corpus Check
- 650 files · ~329,762 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 7507 nodes · 13397 edges · 750 communities (406 shown, 344 thin omitted)
- Extraction: 94% EXTRACTED · 6% INFERRED · 0% AMBIGUOUS · INFERRED: 867 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `6b5f97c2`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- internal/api (service_api.pb.gw.go)
- internal/api (velez_api.pb.gw.go)
- internal/api (matreshka_api.pb.go)
- internal/api (verv_closed_network.pb.gw.go)
- internal/api (control_plane_api.pb.gw.go)
- pkg/web (velez_api.pb.ts)
- pkg/web (ServiceCard.tsx)
- docs/paas (ServiceInfoPage component)
- internal/storage (Storage)
- pkg/web (ServiceInfoPage.tsx)
- docs/review (SCHEDULED_UPGRADE case is empty — upgrades silently do nothing)
- internal/clients (create.go)
- pkg/web (DeployWidget.tsx)
- internal/clients (PortManager)
- internal/storage (storage.go)
- internal/pipelines (fetch_by_api.go)
- internal/transport (Pipeliner)
- internal/transport (grpcServer)
- internal/api (EnumNumber)
- pkg/web (index.ts)
- pkg/web (VervClosedNetworkPage.tsx)
- pkg/web (Router.tsx)
- internal/api (Context)
- pkg/web (service_api.pb.ts)
- internal/clients (cluster.go)
- internal/storage (deployments)
- pkg/web (PluginMatrix.tsx)
- internal/app (custom.go)
- internal/clients (client.go)
- internal/api (MessageState)
- CLAUDE.md (Velez CLAUDE.md project guide)
- internal/storage (node.go)
- internal/service (services.go)
- internal/service (list_test.go)
- internal/transport (deploy_list.go)
- pkg/web (compilerOptions)
- pkg/web (control_plane_api.pb.ts)
- internal/config (launch.go)
- internal/api (VervAppService)
- pkg/web (services.ts)
- internal/cluster (Context)
- internal/pipelines (ConnectServiceToVpn())
- internal/pipelines (ConfigurationService)
- internal/api (Smerd)
- internal/api (CreateSmerd_Request)
- internal/pipelines (LaunchSmerd)
- internal/api (VervPluginType)
- pkg/web (Button.tsx)
- pkg/web (creds.ts)
- internal/pipelines (do_copy_to_volume.go)
- internal/clients (NodeClients)
- internal/pipelines (.UpgradeSmerd())
- internal/api (SubscribeOnChanges_Response)
- tests/e2e (T)
- internal/api (AboutService)
- internal/api (Namespace)
- internal/api (Node)
- pkg/web (devDependencies)
- internal/storage (ListServicesReq)
- internal/api (control_plane_api.pb.go)
- pkg/web (dependencies)
- internal/api (ServiceEnvironmentInfo)
- internal/clients (state.go)
- pkg/web (fetch.pb.ts)
- pkg/web (ServiceService)
- pkg/web (HomePage.tsx)
- internal/storage (docker_impl.go)
- internal/api (ConfigInfo)
- internal/api (VervPlugin_State)
- pkg/web (service.ts)
- pkg/web (verv_closed_network.pb.ts)
- internal/storage (models.go)
- internal/domain (service.go)
- tests/e2e (NewEnvironment())
- internal/api (velez_common.pb.go)
- internal/storage (models.go)
- internal/storage (models.go)
- internal/storage (models.go)
- internal/storage (models.go)
- internal/storage (models.go)
- internal/api (DeploymentInfo)
- internal/pipelines (fetch_from_container.go)
- internal/storage (service_dependencies.go)
- internal/storage (EnvironmentsStorage)
- internal/api (VcnPeer)
- internal/api (ServiceBaseInfo)
- internal/api (Format)
- internal/api (service_api.pb.go)
- internal/api (Message)
- internal/api (MessageState)
- internal/api (ListDeployments_Request)
- internal/api (SizeCache)
- internal/api (UnknownFields)
- internal/api (GetConfig_Request)
- pkg/web (Input.tsx)
- internal/api (Port)
- internal/api (ServiceDependencyInfo)
- internal/cluster (autoupgrade.go)
- internal/service (configurator.go)
- internal/api (Message)
- internal/api (file_service_api_proto_rawDescGZIP())
- internal/api (PatchConfig_Request)
- internal/api (RenameConfig_Request)
- internal/api (SaveConfig_Request)
- pkg/web (control_plane.ts)
- pkg/web (Sidebar.tsx)
- internal/api (BoundResource)
- tests/e2e (helper.go)
- Community 108
- internal/api (Sort_Type)
- internal/api (BreakConnections_Request)
- internal/cluster (local.go)
- internal/api (DeleteConfig_Request)
- pkg/web (SmerdsWidget.tsx)
- Community 114
- internal/api (velez_api.pb.go)
- internal/api (Message)
- internal/clients (Filter)
- internal/cluster (service.go)
- internal/api (CreateConfig_Request)
- pkg/web (DeploymentHistory.tsx)
- internal/api (Plugin)
- internal/api (VervPlugin)
- Community 123
- .claude/factory (factory.sh)
- Community 125
- tests/e2e (helper_environment.go)
- internal/api (ConfigBase)
- docs/paas (Inline-style exception for runtime-parametric atoms)
- tests/e2e (TestEnvironment)
- internal/api (GetHardware_Response)
- internal/transport (.CreateDeploy())
- internal/api (MessageState)
- internal/pipelines (Step)
- internal/storage (New())
- internal/storage (services.go)
- pkg/web (ResourcesSection.tsx)
- internal/api (GetServiceMetrics_Response)
- docs/paas (T38 - PluginManageDialog scaffold (referenced))
- Community 139
- Community 140
- Community 141
- internal/pipelines (create.go)
- internal/api (CreateDeploy_Request)
- internal/api (file_velez_api_proto_rawDescGZIP())
- internal/api (Message)
- internal/api (MessageState)
- internal/api (SizeCache)
- internal/api (UnknownFields)
- internal/api (file_verv_closed_network_proto_rawDescGZIP())
- internal/api (Message)
- internal/transport (network.go)
- pkg/web (ControlPlanePage.tsx)
- internal/api (EnableStatefullCluster)
- Community 154
- Community 155
- internal/transport (VcnNamespace)
- internal/api (ListConfigs_Request)
- internal/api (file_control_plane_api_proto_rawDescGZIP())
- tests/e2e (suite_hello_world_cluster_test.go)
- internal/api (Patch)
- pkg/web (VcnApi)
- pkg/web (compilerOptions)
- internal/api (ListSmerds_Request)
- Community 164
- internal/cluster (Synchronizer)
- internal/pipelines (create_and_inspect.go)
- internal/domain (images.go)
- internal/cluster (DisabledVcnImpl)
- internal/api (DropSmerd_Response)
- internal/clients (api_key_issue.go)
- internal/api (matreshka_common.pb.go)
- internal/api (ListNodes_Response)
- internal/api (verv_closed_network.pb.go)
- internal/storage (service_dependencies.sql.go)
- pkg/web (package.json)
- internal/api (Connection)
- internal/api (Container_Hardware)
- internal/api (Container_Healthcheck)
- internal/api (CreateDeploy_Request_Upgrade_)
- internal/api (Image)
- internal/api (MatreshkaConfigSpec)
- internal/api (NodeBaseInfo)
- internal/api (SearchImageItem)
- internal/api (SearchImages_Request)
- Community 185
- internal/pipelines (get_root_dsn.go)
- internal/patterns (postgres.go)
- internal/domain (vcn.go)
- internal/patterns (headscale.go)
- TestEnvironment
- internal/api (ListNodes_Request)
- internal/api (ListSmerds_Response)
- internal/patterns (.EnableStatefullMode())
- fetchSmerdConfigJob
- internal/api (GetConfigNode_Request)
- internal/api (GetConfigNode_Response)
- UpsertPluginParams
- internal/api (SearchImages_Response)
- internal/api (AssembleConfig_Request)
- internal/api (ConnectService_Request)
- internal/api (ConnectUser_Request)
- internal/api (DropSmerd_Request)
- internal/api (DropSmerd_Response_Error)
- internal/api (EnableHeadscaleServer_DeployHeadscaleConfig)
- internal/api (EnableHeadscaleServer_ExternalHeadscaleConnection)
- internal/api (GetHardware_Response_Value)
- internal/api (UpgradeSmerd_Request)
- .claude/factory (review.sh)
- internal/storage (db.go)
- Community 210
- Community 211
- Version_Response
- internal/clients (.doApiRequest())
- internal/api (file_matreshka_common_proto_rawDescGZIP())
- Context
- T08 — SectionLabel
- internal/storage (db.go)
- internal/storage (db.go)
- internal/storage (db.go)
- internal/storage (db.go)
- internal/storage (service_resources.sql.go)
- internal/storage (db.go)
- pkg/web (ErrorBoundary.tsx)
- internal/api (AssembleConfig_Response)
- internal/api (CreateService_Request)
- internal/api (CreateVcnNamespace_Request)
- internal/api (DeleteVcnNamespace_Request)
- internal/api (GetService_Request)
- internal/api (GetServiceEnvironments_Request)
- internal/api (GetServiceMetrics_Request)
- internal/api (GetServiceResources_Request)
- Message
- internal/api (ListPeers_Request)
- internal/api (StopService_Request)
- internal/api (Version_Response)
- Volume
- tests/e2e (suite_control_plane_test.go)
- internal/clients (list_containers.go)
- internal/clients (Filter)
- Community 240
- selectiveFailDocker
- pkg/web (.GetService())
- pkg/web (MockDeployRow.tsx)
- NewServer
- TasksStorage
- internal/api (CreateVcnNamespace)
- internal/api (DeleteVcnNamespace_Response)
- internal/api (GetHardware_Request)
- internal/api (InitMaster_Response)
- internal/api (ListDeployments)
- internal/api (ListEnvironments)
- Exec
- internal/api (ListVcnNamespaces)
- internal/api (ListVcnNamespaces_Request)
- GetServiceMetrics
- CreateService
- New
- ListNodes
- internal/api (UpgradeSmerd)
- internal/api (UpgradeSmerd_Response)
- createPgUserStep
- .claude/factory (new-task.sh)
- jobs
- internal/service (.ConnectToNetwork())
- internal/service (.DropSmerds())
- internal/transport (.ListEnvironments())
- internal/transport (.ListPlugins())
- ListEnvironments_Response
- internal/clients (hardware_manager.go)
- internal/api (MessageState)
- internal/api (SizeCache)
- internal/api (UnknownFields)
- internal/clients (pull_image.go)
- InitMaster_Response
- VelezAPI
- internal/storage (.GetByName())
- internal/transport (.GetServiceGraph())
- internal/transport (.GetServiceMetrics())
- internal/transport (.GetServiceResources())
- localStorage
- internal/transport (.GetService())
- internal/transport (.RestartService())
- internal/transport (.StopService())
- internal/transport (.GetVervonomicon())
- internal/transport (.ListPeers())
- internal/transport (.DeleteNamespace())
- file_control_plane_api_proto_rawDescGZIP
- portManagerImpl
- internal/transport (.ConnectUser())
- internal/transport (.AssembleConfig())
- internal/transport (.GetHardware())
- internal/transport (.Version())
- internal/transport (.SearchImages())
- internal/transport (.CreateSmerd())
- internal/transport (.DropSmerd())
- internal/transport (.ListSmerds())
- internal/transport (.UpgradeSmerd())
- schemas/Smerds.md (Matreshka (external config service))
- internal/clients (matreshka.go)
- internal/service (.GetServiceEnvironments())
- internal/service (.GetServiceMetrics())
- internal/service (.GetServiceResources())
- internal/service (Context)
- internal/storage (Context)
- start/run_velez.sh (run_velez.sh)
- pluginsStorage
- .claude/factory (status.sh)
- PostgreSQL Database
- CreatePgUserForNode
- internal/clients (.DeleteNamespace())
- internal/clients (.GetNamespace())
- internal/clients (.IssueClientKey())
- internal/clients (.ListNamespaces())
- internal/patterns (portainer.go)
- internal/service (.ListEnvironments())
- Community 318
- Community 319
- Context
- Input.tsx
- ContainerCreateRequest
- pkg/web (TextInput.tsx)
- selectiveFailDocker
- selectiveFailDocker
- internal/app (.InitServers())
- pkg/web (Angie (web server))
- Headscale (VPN / Network Manager)
- Makosh (Velez Ecosystem Component)
- Portainer (Docker Management UI)
- Velez (Node Manager Project)
- internal/storage (querier.go)
- Community 334
- Community 335
- RestartService
- internal/domain (Connection)
- internal/domain (Paging)
- internal/domain (UpgradeSmerd)
- pkg/web (eslint)
- .InspectSmerd
- runner
- internal/storage (querier.go)
- internal/storage (querier.go)
- internal/storage (querier.go)
- internal/storage (querier.go)
- internal/storage (querier.go)
- Community 348
- Community 349
- Community 350
- pkg/web (react)
- pkg/web (react-highlight)
- pkg/web (@tanstack/react-query)
- pkg/web (vite)
- Community 355
- Community 356
- scripts/run_local.sh (run_local.sh)
- ConnectSlave_Response
- Impl
- .enrichServiceWithSmerdData
- internal/clients (ns_list.go)
- .GetServiceEnvironments
- Container_Hardware
- Context
- velez_common.pb.go
- go.mod (go.vervstack.ru/Velez)
- pkg/web (eslint.config.js)
- Community 422
- schemas/Networking.md (Verv Networking 101 (stub))
- Community 428
- Velez UI Redesign — Roadmap
- New
- T30 — AppsPage + AppCard (logical services view)
- EnableStatefullCluster
- ListEnvironments
- Tasks
- FileMountPoint
- .enrichServiceWithSmerdData
- GetConfigNode_Response
- B1-T01 — Extend API for UI wiring (backend)
- Task B2-T01 — Service Runtime Stats API
- Tasks
- Tasks
- Tasks
- Tasks
- Container_Healthcheck
- WatchTask_Request
- GetLoginServerUrl
- ListPlugins_Request
- Project Factory — Agent Context
- Task 000 — Example Task Title
- Task 001 — Health check endpoint
- dockerServiceResourcesStorage
- Task B3-T01 — Plugin Service with Dual-Mode Storage and Hot-Switch
- M2 — Cluster & Networking
- Task 039 — PluginManageDialog: enable action with per-plugin config forms
- T22 — PluginMatrix Widget
- T23 — NetworkTopologyMap Widget
- T25 — MainLayout (rebuild)
- T29 — SearchPage
- Task 032 — PluginMatrix: enable/disable actions and service page navigation
- Task M9-T33 — ServiceInfoPage: Full Redesign
- Task M9-T34 — ObservabilityLinksPanel Widget
- Task M9-T35 — Environment Tab Switcher and Tags Strip
- Task M9-T36 — Sidebar: Fix Active Nav Highlight
- Task M9-T37 — Settings Page
- Task 038 — PluginMatrix: restore status display + separate Manage button with empty dialog
- GetVervonomicon
- VervServicesService
- 🏭 Coding Factory
- Local state
- T02 — StatusDot
- T03 — Badge
- T04 — MiniBar (progress bar)
- T07 — Button (rebuild)
- T09 — StatCard
- T10 — ThreeDotMenu
- T11 — ServiceCard (Kanban card)
- T12 — ServiceListRow (table row)
- T13 — NodeCard
- T14 — VCNPeerRow
- T18 — DeploymentFilters Widget (toolbar)
- T27 — DeploymentsPage (rebuild)
- ListEnvironments
- db.go
- db.go
- WatchTask_Request
- GetLoginServerUrl
- CreateDeploy
- T17 — TopBar Widget
- T19 — KanbanBoard Widget
- T20 — ServiceListView Widget
- T21 — NodeHealthList Widget
- T24 — VCNPeerTable Widget
- T26 — ControlPlanePage (rebuild)
- T28 — VCNPage (rebuild)
- DeleteEnvironment
- GetServiceResources
- fetch_by_api.go
- VervServicesService
- renameContainerStep
- initKey
- CreateSmerdTaskPayload
- factory-worker.md
- Smerds management logic
- Velez (lightweight node manager)
- DataSourcesConfig
- React + TypeScript + Vite
- create_service.go
- easyp.yaml proto codegen config
- VCN Headscale Config
- querier.go
- querier.go
- Networking.md
- factory-worker agent
- ListVervServices API process
- FetchSmerds API process
- AppCard component (AppData type)
- ListNodes
- IconButton component
- StatCard component
- StatusDot component
- Chip component
- EnvChip component
- IncidentChip component
- Project Factory Agent Context
- Coding Factory pipeline README
- Velez CLAUDE.md project guide
- factory skill (Haiku→Ollama pipeline)
- task skill (create task from description)
- CodeBlock component
- Coding factory pipeline architecture
- Frontend code style rules (named functions, CSS Modules, etc.)
- Go error handling rule (no assign-in-if)
- Go struct literal rule (named variable first)
- Claude Haiku review step
- Pi Ollama code-generation host
- PatchConfig_Response
- Option B: shell factory with direct Ollama API calls
- Config subscription is commented out (handleConfigurationSubscription)
- GetService_Request
- internal/service/service_manager/container_manager/smerd_list.go
- internal/service/service_manager/container_manager/smerds_drop.go
- ControlPlanePage component
- NodeHealthList widget
- PluginMatrix widget
- createSmerdStep
- .CreateNamespace
- ListNodes
- createPgUserStep
- DeployPage (existing deploy form)
- UpdateEnvironment
- DeploymentFilters widget
- DeploymentsPage component
- KanbanBoard widget
- ServiceListView widget
- DeploymentsSection component
- Pure UI component must not call API directly (callback-only pattern)
- ApplyNetworkPolicy (pipeline step)
- ApplyWebserverConfig (pipeline step)
- auth.yaml (inter-service access control)
- configuration.yaml (app config)
- CreateService (vervonomicon pipeline step)
- dependencies.yaml (external resources)
- Vervonomicon (declarative deployment descriptor)
- vervonomicon.yaml (index file)
- network.yaml (ports/exposure)
- ProvisionDependencies (pipeline step)
- RegisterSecrets (pipeline step)
- resources.yaml (hardware/placement)
- secrets.yaml (secret declarations)
- SyncConfiguration (pipeline step)
- webserver.conf (nginx fragment)
- ListNodes RPC
- ListPeers RPC (VCN)
- ServiceBaseInfo (proto message)
- B1-T01: Extend API for UI wiring
- VcnPeer (proto message)
- Docker ContainerStats client (internal/clients/docker)
- GetServiceStats RPC
- B2-T01: Service Runtime Stats API
- ClusterStateManagerContainer (existing hot-switch pattern)
- ListPlugins RPC
- ListVervServices handler (list_verv_services.go)
- PluginService interface
- PluginServiceContainer (atomic.Pointer hot-switch)
- B3-T01: Plugin Service with Dual-Mode Storage and Hot-Switch
- ListServices (fetched by HomePage)
- RestartService endpoint (missing from gRPC API)
- StopService endpoint (missing from gRPC API)
- CreateNewDeployment API call
- DeploymentWidget component
- DeployMenu component
- ListDeployments (filtered by service key)
- CreateService call (NewServicePage)
- InitServiceWidget
- NewServicePage
- VITE_VELEZ_BACKEND_URL / VITE_VELEZ_AUTH_HEADER build-time env vars
- catchGrpc (API error toast handler)
- ErrorBoundary.tsx
- cpu-heavy tag (auto-assigned)
- disk-heavy tag (auto-assigned)
- Node Scheduling Tags
- short-living tag
- stateful-priority tag
- test-only tag
- Role-based access control (admin/deployer/viewer)
- Auto-rollback on failed health check
- Health-gated promotion
- Version
- TestMain
- Context
- InitMaster
- .EnablePlugin
- ListPlugins
- .DeleteEnvironment
- pre-commit
- CacheKey.Plugins query cache key
- ControlPlanePage
- EnableHeadscaleServer payload type
- EnableService process function / RPC
- EnableStatefullCluster payload type
- EnableStatefullPgCluster process function
- PluginManageDialog component
- T39 - PluginManageDialog enable action with per-plugin config forms
- VervServiceType (proto enum)
- T02 - StatusDot task spec
- Badge component
- T03 - Badge task spec
- MiniBar component
- T04 - MiniBar task spec
- T05 - Status Chips task spec
- ActionButton component (existing, distinct purpose)
- IconButton component
- T06 - IconButton task spec
- T07 - Button rebuild task spec
- T08 - SectionLabel task spec
- StatCard component
- T09 - StatCard task spec
- T10 - ThreeDotMenu task spec
- ThreeDotMenu component
- ServiceCard component
- T11 - ServiceCard (Kanban card) task spec
- ServiceListRow component
- T12 - ServiceListRow task spec
- NodeCard component
- T13 - NodeCard task spec
- T14 - VCNPeerRow task spec
- VCNPeerRow component
- T15 - CodeBlock task spec
- T16 - Sidebar Widget task spec
- T17 - TopBar Widget task spec
- TopBar widget
- DeploymentFilters widget
- T18 - DeploymentFilters Widget task spec
- KanbanBoard widget
- T19 - KanbanBoard Widget task spec
- ServiceListView widget
- T20 - ServiceListView Widget task spec
- NodeHealthList widget
- T21 - NodeHealthList Widget task spec
- PluginMatrix widget
- T22 - PluginMatrix Widget task spec
- NetworkTopologyMap widget
- T23 - NetworkTopologyMap Widget task spec
- T24 - VCNPeerTable Widget task spec
- VCNPeerTable widget
- ListSmerds / cluster nodes endpoint
- MainLayout (rebuild)
- PageHeader (old top-navbar component, removed)
- Router / Routes enum
- SettingsWidget (referenced, ambiguous placement)
- T25 - MainLayout rebuild task spec
- Toaster component
- fakeJobsEngine
- ListDeployments
- CreateDeploy
- ListVervPlugins
- InitMaster_Response
- golangci-lint configuration
- StopService_Response
- clients.go
- .resolveEnvironment
- local.issues.md — known code issues
- smerds.ts mapping module (mapSmerdToServiceCard, mapSmerdToAppData, mapSmerdStatus, LOCAL_NODE)
- .RegisterNode
- AddMakoshRecord
- NodeCard component (NodeCardData type)
- DeployMenu component
- Service page Header component
- Velez-UI Layer Structure Rule
- Velez-UI Styling Rules
- Velez-UI Toast/Error Handling
- useToaster Zustand Store
- VervClosedNetworkPage (/vcn route)
- moti.yaml Proto Codegen Config
- Velez-UI Vite+React Template README
- PluginManageDialog component
- Velez project README
- MainLayout component
- Router.tsx
- Routes.ts route enum
- Routes.Search route constant
- ContainerApi (Docker engine abstraction)
- Smerd Update/Restart Flow
- SearchPage component
- ObservabilityLinksPanel widget
- ServiceCard component (ServiceCardData type)
- ServiceInfoPage component
- useCredentialsStore (creds.ts)
- SettingsPage component
- SettingsWidget component (modal)
- useSettings.ts hook (backendUrl, authHeader, observability URLs)
- useSettings hook (initReq)
- Sidebar active nav highlight bug (activeNodeId vs activeNav)
- Sidebar widget
- SmerdMetaSection component
- sqlc.yaml query generation config
- T38 - PluginManageDialog scaffold (referenced)
- Task file template (000)
- StepsDialogHeader.tsx
- createPgUserStep
- NetworkTopologyMap widget
- VCNPeerRow (VCNPeerData type)
- VCNPeerTable widget
- VervClosedNetworkPage component
- dedicatedStorage
- ListPlugins_Request
- RemoveService
- Jobs Migration — Open Questions
- UpdateEnvironment
- drop_smerd.go
- ListPlugins_Request
- Configurator
- Pipelines → Jobs Migration
- EnableStatefullTaskPayload
- VPN client key (Headscale auth key) storage in task context

## God Nodes (most connected - your core abstractions)
1. `T` - 401 edges
2. `newFakeDocker()` - 73 edges
3. `NodeClients` - 63 edges
4. `file_service_api_proto_rawDescGZIP()` - 50 edges
5. `newFakeNodeClients()` - 48 edges
6. `useToaster` - 45 edges
7. `newFakeContainerAPI()` - 41 edges
8. `file_control_plane_api_proto_rawDescGZIP()` - 40 edges
9. `NewEnvironment()` - 37 edges
10. `proto()` - 34 edges

## Surprising Connections (you probably didn't know these)
- `moti.yaml proto codegen config` --semantically_similar_to--> `easyp.yaml proto codegen config`  [INFERRED] [semantically similar]
  moti.yaml → easyp.yaml
- `b64Decode()` --references--> `T`  [EXTRACTED]
  pkg/web/Velez-UI/src/app/api/velez/fetch.pb.ts → internal/clients/cluster_clients/headscale/client_key_read.go
- `WithMatreshka()` --calls--> `WithSharedInstance()`  [INFERRED]
  tests/e2e/helper_environment.go → internal/cluster/configuration/service.go
- `initConfig()` --calls--> `Load()`  [INFERRED]
  tests/e2e/helper_environment.go → internal/config/load.go
- `Matreshka Service Icon` --references--> `matreshka (external configuration service)`  [EXTRACTED]
  pkg/web/Velez-UI/src/assets/icons/services/matreshka.png → CLAUDE.md

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Velez Configuration Environment Variable Schema** — config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_configyaml, config_configtemplate, docs_configuration [INFERRED 0.90]
- **Matreshka Config Mock Pattern** — tests_config_mocks_hello_world_matreshkaconfig, tests_config_mocks_velez_default_config_matreshkaconfig, schemas_smerds_matreshka [INFERRED 0.75]

## Communities (750 total, 344 thin omitted)

### Community 0 - "internal/api (service_api.pb.gw.go)"
Cohesion: 0.06
Nodes (87): ClientConnInterface, Context, CreateDeploy_Request, CreateDeploy_Response, CreateService_Request, CreateService_Response, GetService_Request, GetService_Response (+79 more)

### Community 1 - "internal/api (velez_api.pb.gw.go)"
Cohesion: 0.07
Nodes (72): AssembleConfig_Request, AssembleConfig_Response, BreakConnections_Request, BreakConnections_Response, ClientConnInterface, Context, CreateSmerd_Request, DropSmerd_Request (+64 more)

### Community 2 - "internal/api (matreshka_api.pb.go)"
Cohesion: 0.04
Nodes (21): file_matreshka_api_proto_init(), file_matreshka_api_proto_rawDescGZIP(), Message, MessageState, SizeCache, UnknownFields, init(), CreateConfig (+13 more)

### Community 3 - "internal/api (verv_closed_network.pb.gw.go)"
Cohesion: 0.09
Nodes (52): ClientConnInterface, ConnectService_Request, ConnectService_Response, ConnectUser_Request, ConnectUser_Response, Context, CreateVcnNamespace_Request, CreateVcnNamespace_Response (+44 more)

### Community 4 - "internal/api (control_plane_api.pb.gw.go)"
Cohesion: 0.08
Nodes (62): ConnectSlave_Request, ConnectSlave_Response, _ControlPlaneAPI_ConnectSlave_Handler(), _ControlPlaneAPI_CreateEnvironment_Handler(), _ControlPlaneAPI_DeleteEnvironment_Handler(), _ControlPlaneAPI_EnablePlugin_Handler(), _ControlPlaneAPI_ListEnvironments_Handler(), _ControlPlaneAPI_ListNodes_Handler() (+54 more)

### Community 5 - "pkg/web (velez_api.pb.ts)"
Cohesion: 0.12
Nodes (24): RuntimeResolver, connectToNetwork(), disconnectFromNetworks(), fromContainerNetwork(), ContainerState, Context, EndpointSettings, InspectResponse (+16 more)

### Community 6 - "pkg/web (ServiceCard.tsx)"
Cohesion: 0.07
Nodes (43): Smerd, SmerdStatus, AppCard(), AppCardProps, EnvChip(), EnvChipProps, FreezeChip(), IncidentChip() (+35 more)

### Community 7 - "docs/paas (ServiceInfoPage component)"
Cohesion: 0.14
Nodes (13): 1. DeploymentsPage (`src/pages/deployments/DeploymentsPage.tsx`), 2. AppsPage (`src/pages/apps/AppsPage.tsx`), 3. SearchPage (`src/pages/search/SearchPage.tsx`), 4. ControlPlanePage (`src/pages/controlplane/ControlPlanePage.tsx`), Acceptance Criteria, Blocked dependency, Do NOT change, Files to create (+5 more)

### Community 8 - "internal/storage (Storage)"
Cohesion: 0.09
Nodes (4): file_service_api_proto_rawDescGZIP(), CreateDeploy_Response, CreateService_Response, GetServiceEnvironments_Request

### Community 9 - "pkg/web (ServiceInfoPage.tsx)"
Cohesion: 0.11
Nodes (4): MessageState, CreateService, GetServiceGraph, GetServiceResources_Request

### Community 12 - "pkg/web (DeployWidget.tsx)"
Cohesion: 0.13
Nodes (18): classifyImage(), fromMatreshkaYamlToEvon(), fromYamlToEvon(), APIClient, ConfigFormat, Context, Docker, isPostgresByImageTags() (+10 more)

### Community 13 - "internal/clients (PortManager)"
Cohesion: 0.12
Nodes (14): PluginBaseInfo, DeploymentStatus, VervPlugin_State, Context, Docker, newPluginsStorage(), Context, NullInt64 (+6 more)

### Community 14 - "internal/storage (storage.go)"
Cohesion: 0.05
Nodes (19): staticProvider, Pointer, Storage, Querier, envProvider, fakeClusterStorage, localStorage, Container (+11 more)

### Community 15 - "internal/pipelines (fetch_by_api.go)"
Cohesion: 0.12
Nodes (19): Environment, CreateServiceRequest, DeploymentInfo, EnvironmentStore, NOTE: this is unrelated to `ServiceEnvironment`, EnvironmentDeleteDialog(), EnvironmentDeleteDialogProps, EnvironmentManageDialogProps (+11 more)

### Community 16 - "internal/transport (Pipeliner)"
Cohesion: 0.14
Nodes (16): FromVolume(), Container_Settings, Volume, ToVolume(), Context, Docker, Volume, writeFileToContainer() (+8 more)

### Community 17 - "internal/transport (grpcServer)"
Cohesion: 0.05
Nodes (61): EnableStatefullSuite, T, TestCreateSmerdRequest_NilVervAndPlainStayNil(), TestCreateSmerdRequest_PlainSurvivesRoundTrip(), TestCreateSmerdRequest_SurvivesRoundTripEmbeddedInTaskPayload(), TestCreateSmerdRequest_VervAndPlainBothSurviveRoundTrip(), TestCreateSmerdRequest_VervSurvivesRoundTrip(), TestUpgradeSmerdRequest_SurvivesRoundTripEmbeddedInTaskPayload() (+53 more)

### Community 18 - "internal/api (EnumNumber)"
Cohesion: 0.08
Nodes (7): EnumDescriptor, EnumNumber, EnumType, ConfigFormat, NodeStatus, RestartPolicyType, Smerd_Status

### Community 19 - "pkg/web (index.ts)"
Cohesion: 0.08
Nodes (53): CopyToVolumeTaskPayload, TestCreateScratchContainerJob_ContainerCreateError(), TestCreateScratchContainerJob_Rollback_NoContainerId_NoOp(), TestCreateScratchContainerJob_Rollback_RemoveErrorPropagates(), TestCreateScratchContainerJob_Rollback_RemovesContainer(), TestCreateSidecarContainerJob_ContainerCreateError(), TestCreateSidecarContainerJob_Rollback_NoContainerId_NoOp(), TestCreateSidecarContainerJob_Rollback_NotFoundIsSwallowed() (+45 more)

### Community 20 - "pkg/web (VervClosedNetworkPage.tsx)"
Cohesion: 0.14
Nodes (25): ConnectServiceToVpnTaskPayload, NewConnectServiceToVpnHandler(), connectServiceToVpnTask(), VelezTask, patchStartAPIField(), TestAddMakoshRecordJob_Success(), TestAddMakoshRecordJob_UpsertError(), TestConnectServiceToVpnHandler_Action() (+17 more)

### Community 21 - "pkg/web (Router.tsx)"
Cohesion: 0.05
Nodes (43): NodeStatus, useEnvironmentStore, MainLayout(), NAV_TO_ROUTE, NavId, ROUTE_TO_NAV, TOOL_TO_ROUTE, ToolId (+35 more)

### Community 22 - "internal/api (Context)"
Cohesion: 0.09
Nodes (23): BidiStreamingClient, ClientConnInterface, Context, ServiceRegistrar, UnaryServerInterceptor, Version_Request, Version_Response, _MatreshkaApi_CreateConfig_Handler() (+15 more)

### Community 23 - "pkg/web (service_api.pb.ts)"
Cohesion: 0.03
Nodes (61): AboutService, Absent, BaseCreateDeployRequest, BaseGetServiceResponse, BoundResource, CreateDeploy, CreateDeployRequest, CreateDeployRequestUpgrade (+53 more)

### Community 24 - "internal/clients (cluster.go)"
Cohesion: 0.03
Nodes (79): AssembleConfig, AssembleConfigRequest, AssembleConfigResponse, BreakConnections, BreakConnectionsRequest, BreakConnectionsResponse, CreateSmerd, CreateSmerdRequest (+71 more)

### Community 25 - "internal/storage (deployments)"
Cohesion: 0.10
Nodes (19): ServiceBaseInfo, TestToServiceBaseInfoList(), TestToServiceBaseInfoWithEmptyFields(), TestToServiceBaseInfoWithEnrichedFields(), toServiceBaseInfo(), toServiceBaseInfoList(), Context, GetServiceEnvironments_Request (+11 more)

### Community 26 - "pkg/web (PluginMatrix.tsx)"
Cohesion: 0.07
Nodes (20): AssembleConfigJobSuite, AssembleConfigSuite, HelloWorldClusterSuite, Context, runner[T], Test_Lifecycle(), Context, Suite (+12 more)

### Community 27 - "internal/app (custom.go)"
Cohesion: 0.04
Nodes (56): ListNodesResponse, ConfigFormat, Connection, Container, ContainerHardware, ContainerHealthcheck, ContainerSettings, FileConfig (+48 more)

### Community 28 - "internal/clients (client.go)"
Cohesion: 0.12
Nodes (13): Docker, BoundResource, ContainerStats, ServiceMetrics, asciiSymbolsOnly(), APIClient, Context, ExecOptions (+5 more)

### Community 29 - "internal/api (MessageState)"
Cohesion: 0.22
Nodes (5): ClusterStateManager, Storage, Pointer, NewContainer(), stateManager

### Community 30 - "CLAUDE.md (Velez CLAUDE.md project guide)"
Cohesion: 0.40
Nodes (6): makosh/headscale (VPN/network management), matreshka (external configuration service), Velez local state (VelezKey / local_state.json), config/config.yaml (production config), config/config_template.yaml, Matreshka Service Icon

### Community 31 - "internal/storage (node.go)"
Cohesion: 0.10
Nodes (22): ConnectionHealth, ConnectionHealthStatus, useConnectionHealth(), Credentials, ls, useCredentialsStore, InitReq, Settings (+14 more)

### Community 32 - "internal/service (services.go)"
Cohesion: 0.08
Nodes (23): 1. Ubuntu / Debian (apt), 2. Homebrew (macOS), 3. Chocolatey (Windows), Community repo vs. self-hosted feed, Config file placement, Docker's host-access requirements don't disappear on macOS, Does Velez support running as a native Windows service today?, Formula structure and the goreleaser nuance (+15 more)

### Community 33 - "internal/service (list_test.go)"
Cohesion: 0.11
Nodes (16): FS, HandlerFunc, App, ServerOption, GrpcIncomingInterceptor(), ServerOption, LogInterceptor(), ServerOption (+8 more)

### Community 34 - "internal/transport (deploy_list.go)"
Cohesion: 0.09
Nodes (24): NodeBaseInfo, NodeStatusFromLastOnline(), Test_NodeStatusFromLastOnline_DisabledNodeIsOffline(), Test_NodeStatusFromLastOnline_FreshHeartbeatIsOnline(), Test_NodeStatusFromLastOnline_PastOfflineThresholdIsOffline(), Test_NodeStatusFromLastOnline_StaleUnderOfflineThresholdIsDegraded(), Test_ToBasicNodeInfo_PassesThroughRegionUnchanged(), Test_ToBasicNodeInfo_PassesThroughUsageFieldsUnchanged() (+16 more)

### Community 35 - "pkg/web (compilerOptions)"
Cohesion: 0.08
Nodes (24): compilerOptions, allowImportingTsExtensions, allowJs, allowSyntheticDefaultImports, esModuleInterop, forceConsistentCasingInFileNames, isolatedModules, jsx (+16 more)

### Community 37 - "internal/config (launch.go)"
Cohesion: 0.23
Nodes (9): fetchConfigStep, FetchConfig(), Context, FileConfig, InspectResponse, MatreshkaConfigSpec, isPostgresByImageTags(), isVervImage() (+1 more)

### Community 38 - "internal/api (VervAppService)"
Cohesion: 0.08
Nodes (7): EnumDescriptor, EnumNumber, EnumType, DeploymentStatus, GetService_Response_VervService, NodeType, VervAppService

### Community 39 - "pkg/web (services.ts)"
Cohesion: 0.12
Nodes (13): VelezTask, Context, NullRawMessage, NullString, NullTime, Queries, VelezTask, VelezTaskStatus (+5 more)

### Community 40 - "internal/cluster (Context)"
Cohesion: 0.08
Nodes (15): ApiVersion_Request, ApiVersion_Response, disabledConfigurator, disabledServiceDiscovery, BidiStreamingClient, Context, ListEndpoints_Request, ListEndpoints_Response (+7 more)

### Community 41 - "internal/pipelines (ConnectServiceToVpn())"
Cohesion: 0.10
Nodes (38): TestFetchSmerdConfigJob_SetEnv_CallerValueNotOverwritten(), TestFetchSmerdConfigJob_VervAndPlain_BothApply(), newFakeConfigurationService(), newFakeContainerService(), newFakeRuntimes(), NewUpgradeSmerdHandler(), VelezTask, realPortManager() (+30 more)

### Community 42 - "internal/pipelines (ConfigurationService)"
Cohesion: 0.06
Nodes (36): AssembleConfigTaskPayload, ConnectServiceToVpnTaskPayload, CopyToVolumeTaskPayload, CreateServiceTaskPayload, CreateSmerdTaskPayload, DropSmerdTaskPayload, EnableStatefullTaskPayload, TasksApi (+28 more)

### Community 43 - "internal/api (Smerd)"
Cohesion: 0.11
Nodes (3): Timestamp, NetworkBind, Smerd

### Community 44 - "internal/api (CreateSmerd_Request)"
Cohesion: 0.10
Nodes (7): Container_Hardware, Container_Healthcheck, Container_Settings, FileConfig, MatreshkaConfigSpec, RestartPolicy, CreateSmerd_Request

### Community 45 - "internal/pipelines (LaunchSmerd)"
Cohesion: 0.10
Nodes (18): Context, CreateRequest, Docker, ServiceDiscovery, CreateRequest, TailScaleContainerSidecar(), addMakoshRecordJob, clientKeyAccessor (+10 more)

### Community 46 - "internal/api (VervPluginType)"
Cohesion: 0.31
Nodes (7): Context, ListPlugins_Response, Storage, listInactivePlugins(), NewPluginService(), Plugin, pluginService

### Community 47 - "pkg/web (Button.tsx)"
Cohesion: 0.09
Nodes (25): NullFloat64, NullInt32, NullInt64, NullRawMessage, NullString, NullTime, Time, Value (+17 more)

### Community 48 - "pkg/web (creds.ts)"
Cohesion: 0.10
Nodes (18): Common vs. divergent implementation (embedding), Interface design, Names are always virtual at the interface boundary, Naming, Resolution and hot-swap, Suffix filtering has no "unscoped" escape hatch, The interface, Wrapper types for Docker SDK parameters (+10 more)

### Community 49 - "internal/pipelines (do_copy_to_volume.go)"
Cohesion: 0.05
Nodes (39): b64, b64Decode(), b64Encode(), fetchStreamingRequest(), FlattenedRequestPayload, flattenRequestPayload(), getNewLineDelimitedJSONDecodingStream(), getNotifyEntityArrivalSink() (+31 more)

### Community 50 - "internal/clients (NodeClients)"
Cohesion: 0.17
Nodes (32): newCreateRequest(), newFakeInspectOnlyAPI(), newFakeRemoveAPI(), newFakeRenameAPI(), newInspectResponse(), newInspectResponseWithState(), newLabelRuntime(), TestLabelBasedRuntime_EmptySuffixLeavesNameUntouched() (+24 more)

### Community 51 - "internal/pipelines (.UpgradeSmerd())"
Cohesion: 0.12
Nodes (26): fakeRuntimeResolver, dedicatedStorage, EnvironmentsProvider, resolver, APIClient, NewResolver(), Context, Environment (+18 more)

### Community 52 - "internal/api (SubscribeOnChanges_Response)"
Cohesion: 0.17
Nodes (4): BidiStreamingServer, ServerStream, _MatreshkaApi_SubscribeOnChanges_Handler(), SubscribeOnChanges_Request

### Community 53 - "tests/e2e (T)"
Cohesion: 0.43
Nodes (7): ConnectToNetworkRequest, ConnectToNetwork(), CreateNetwork(), DisconnectFromNetworks(), APIClient, Context, EndpointSettings

### Community 54 - "internal/api (AboutService)"
Cohesion: 0.09
Nodes (3): AboutService, GetService_Response, isGetService_Response_Payload

### Community 55 - "internal/api (Namespace)"
Cohesion: 0.10
Nodes (3): CreateVcnNamespace_Response, ListVcnNamespaces_Response, Namespace

### Community 57 - "pkg/web (devDependencies)"
Cohesion: 0.09
Nodes (23): devDependencies, eslint, eslint-import-resolver-typescript, @eslint/js, eslint-plugin-import-x, eslint-plugin-react, eslint-plugin-react-hooks, eslint-plugin-react-refresh (+15 more)

### Community 58 - "internal/storage (ListServicesReq)"
Cohesion: 0.10
Nodes (17): Context, Queries, newServiceDependenciesStorage(), BoundResource, Context, countTotal(), fromStorageToDomainService(), Context (+9 more)

### Community 59 - "internal/api (control_plane_api.pb.go)"
Cohesion: 0.08
Nodes (11): EnableHeadscaleServer_DeployHeadscaleConfig, EnableHeadscaleServer_ExternalHeadscaleConnection, file_control_plane_api_proto_init(), init(), EnableHeadscaleServer, EnableHeadscaleServer_DeployConfig, EnableHeadscaleServer_ExternalConnect, EnablePlugin_Request (+3 more)

### Community 60 - "pkg/web (dependencies)"
Cohesion: 0.09
Nodes (25): NullFloat64, NullInt32, NullInt64, NullRawMessage, NullString, NullTime, Time, Value (+17 more)

### Community 61 - "internal/api (ServiceEnvironmentInfo)"
Cohesion: 0.10
Nodes (3): Timestamp, GetServiceEnvironments_Response, ServiceEnvironmentInfo

### Community 62 - "internal/clients (state.go)"
Cohesion: 0.12
Nodes (11): firstNotEmptyKey(), Once, RWMutex, readStateFromPath(), writeKey(), fakeStateManager, ClusterState, Headscale (+3 more)

### Community 63 - "pkg/web (fetch.pb.ts)"
Cohesion: 0.07
Nodes (22): ConnectServiceToVcn, GetVcnAuthKeyReq, IssueClientKey, ListVpnNamespaces, RegisterVcnNodeReq, VcnAuthKey, VcnNamespace, createNamespaceRequest (+14 more)

### Community 64 - "pkg/web (ServiceService)"
Cohesion: 0.08
Nodes (28): EnvCard(), EnvCardProps, ServiceEnvironment, ServiceGraphNode, ServiceResource, VervonomiconDocs, useGetServiceGraphQuery(), useGetServiceResourcesQuery() (+20 more)

### Community 65 - "pkg/web (HomePage.tsx)"
Cohesion: 0.09
Nodes (25): NullVelezDeploymentStatus, NullVelezJobStatus, NullVelezTaskStatus, VelezDeployment, VelezDeploymentSpecification, VelezDeploymentStatus, VelezEnvironment, VelezJob (+17 more)

### Community 66 - "internal/storage (docker_impl.go)"
Cohesion: 0.11
Nodes (3): Message, GetService_Request, GetVervonomicon_Request

### Community 67 - "internal/api (ConfigInfo)"
Cohesion: 0.04
Nodes (65): DeploymentStatus, ServiceBaseInfo, Toast, Toaster, useToaster, TagChip(), TagChipProps, SkeletonLoader() (+57 more)

### Community 68 - "internal/api (VervPlugin_State)"
Cohesion: 0.08
Nodes (7): EnumDescriptor, EnumNumber, EnumType, Plugin, VervPlugin, VervPlugin_State, VervPluginType

### Community 69 - "pkg/web (service.ts)"
Cohesion: 0.09
Nodes (16): FinalStepProps, Step, StepScreenProps, StepsDialog(), StepsDialogProps, TestContext, StepsDialogHeader(), StepsDialogHeaderContent (+8 more)

### Community 70 - "pkg/web (verv_closed_network.pb.ts)"
Cohesion: 0.08
Nodes (52): Code, RespError, PgName(), TestPgName_EmptySuffix_ReturnsBaseName(), TestPgName_WithSuffix_AppendsSuffix(), proto(), TestStartSidecarContainerJob_ContainerStartError(), TestStartSidecarContainerJob_NoContainerId_Error() (+44 more)

### Community 71 - "internal/storage (models.go)"
Cohesion: 0.09
Nodes (25): NullVelezDeploymentStatus, NullVelezJobStatus, NullVelezTaskStatus, VelezDeployment, VelezDeploymentSpecification, VelezDeploymentStatus, VelezEnvironment, VelezJob (+17 more)

### Community 72 - "internal/domain (service.go)"
Cohesion: 0.17
Nodes (10): Context, Duration, VelezTask, APIClient, Context, Handler, ServiceRegistrar, Impl (+2 more)

### Community 73 - "tests/e2e (NewEnvironment())"
Cohesion: 0.04
Nodes (51): containerRuntimeTestCase, ControlPlaneSuite, EnvironmentsSuite, LifecycleSuite, StateOpt, TestEnvironment, TestEnvOpt, UpgradeSmerdSuite (+43 more)

### Community 74 - "internal/api (velez_common.pb.go)"
Cohesion: 0.53
Nodes (4): keyIssuer, Context, Docker, issueNewAPIKey()

### Community 75 - "internal/storage (models.go)"
Cohesion: 0.09
Nodes (25): NullFloat64, NullInt32, NullInt64, NullRawMessage, NullString, NullTime, Time, Value (+17 more)

### Community 76 - "internal/storage (models.go)"
Cohesion: 0.09
Nodes (25): NullFloat64, NullInt32, NullInt64, NullRawMessage, NullString, NullTime, Time, Value (+17 more)

### Community 77 - "internal/storage (models.go)"
Cohesion: 0.09
Nodes (25): NullFloat64, NullInt32, NullInt64, NullRawMessage, NullString, NullTime, Time, Value (+17 more)

### Community 78 - "internal/storage (models.go)"
Cohesion: 0.09
Nodes (25): NullFloat64, NullInt32, NullInt64, NullRawMessage, NullString, NullTime, Time, Value (+17 more)

### Community 79 - "internal/storage (models.go)"
Cohesion: 0.09
Nodes (25): NullFloat64, NullInt32, NullInt64, NullRawMessage, NullString, NullTime, Time, Value (+17 more)

### Community 81 - "internal/pipelines (fetch_from_container.go)"
Cohesion: 0.07
Nodes (32): fakeVervServices, CreateEnvironmentReq, DeleteEnvironmentReq, Environment, UpdateEnvironmentReq, pgStorage, staticStorage, Time (+24 more)

### Community 82 - "internal/storage (service_dependencies.go)"
Cohesion: 0.14
Nodes (8): Docker, Mutex, Service, fakeContainerRuntime, fakeNodesStorage, fakeRuntimeRenameCall, fakeRuntimeResolver, fakeServicesStorage

### Community 83 - "internal/storage (EnvironmentsStorage)"
Cohesion: 0.17
Nodes (11): LaunchSmerd, LaunchSmerdResult, CreateSmerd_Request, pipeliner, APIClient, Context, Healthcheck(), Context (+3 more)

### Community 86 - "internal/api (Format)"
Cohesion: 0.21
Nodes (9): AutoUpgrade, Tx, APIClient, Context, Duration, Once, Summary, imageNameWithoutTag() (+1 more)

### Community 87 - "internal/api (service_api.pb.go)"
Cohesion: 0.23
Nodes (7): Context, Mutex, Queries, Tx, VelezTask, newTasksStorage(), tasks

### Community 88 - "internal/api (Message)"
Cohesion: 0.09
Nodes (25): ClusterStateManagerContainer, getExposedPgPort(), Container, Context, InspectResponse, CreatePgUserForNode(), Step, containerIDAccessor (+17 more)

### Community 89 - "internal/api (MessageState)"
Cohesion: 0.25
Nodes (6): CreateSmerd_Request, createNetworkIfMissing(), CreateSmerd_Request, createNetworkAPI, prepareSmerdVervConfigJob, prepareUpgradeVervConfigJob

### Community 90 - "internal/api (ListDeployments_Request)"
Cohesion: 0.13
Nodes (3): Paging, ListDeployments_Request, ListServices_Request

### Community 91 - "internal/api (SizeCache)"
Cohesion: 0.11
Nodes (14): Context, Impl, CreateEnvironment_Request, CreateEnvironment_Response, environmentToPb(), Context, Impl, Environment (+6 more)

### Community 92 - "internal/api (UnknownFields)"
Cohesion: 0.31
Nodes (13): availablePortsForTest(), TestContainer_ForwardsEnvironmentScopedAPI(), TestGetPortForEnvironment_RecordsOwner(), TestLockPort_UnscopedStillCollidesWithScoped(), TestLockPortForEnvironment_CrossEnvironmentCollisionRejected(), TestLockPortForEnvironment_RecordsOwner(), TestLockPortForEnvironment_RollsBackOwnershipOnFailure(), TestLockPortForEnvironment_SameEnvironmentDoubleLockRejected() (+5 more)

### Community 93 - "internal/api (GetConfig_Request)"
Cohesion: 0.15
Nodes (11): Context, Impl, EnablePlugin_Request, EnablePlugin_Response, Context, Impl, Handler, ServiceRegistrar (+3 more)

### Community 94 - "pkg/web (Input.tsx)"
Cohesion: 0.31
Nodes (5): APIClient, Context, InspectResponse, PrepareVervConfig(), prepareVervConfig

### Community 98 - "internal/service (configurator.go)"
Cohesion: 0.20
Nodes (12): VelezJobStatus, TaskStatus_JobStatus, VelezJobStatus, VelezTask, VelezTaskStatus, jobStatusesToProto(), jobStatusToProto(), taskStatusToProto() (+4 more)

### Community 101 - "internal/api (PatchConfig_Request)"
Cohesion: 0.22
Nodes (7): ListEndpoints_Request, ListEndpoints_Response, UpsertEndpoints_Request, UpsertEndpoints_Response, Version_Request, Version_Response, fakeServiceDiscovery

### Community 103 - "internal/api (SaveConfig_Request)"
Cohesion: 0.07
Nodes (24): CopyToContainerOptions, CreateOptions, Context, APIClient, Context, CreateResponse, EndpointSettings, ExecOptions (+16 more)

### Community 104 - "pkg/web (control_plane.ts)"
Cohesion: 0.12
Nodes (14): CopyToVolumeRequest, pipeliner, NewCopyToVolumeRunner(), pipeliner, Runner, DropContainerStep(), Context, Docker (+6 more)

### Community 105 - "pkg/web (Sidebar.tsx)"
Cohesion: 0.05
Nodes (16): file_matreshka_common_proto_init(), file_matreshka_common_proto_rawDescGZIP(), Message, MessageState, SizeCache, Timestamp, UnknownFields, init() (+8 more)

### Community 107 - "tests/e2e (helper.go)"
Cohesion: 0.23
Nodes (5): Context, Queries, Tx, newDeploymentsStorage(), deployments

### Community 109 - "internal/api (Sort_Type)"
Cohesion: 0.11
Nodes (5): EnumDescriptor, EnumNumber, EnumType, ConfigType, Sort_Type

### Community 110 - "internal/api (BreakConnections_Request)"
Cohesion: 0.15
Nodes (3): Connection, BreakConnections_Request, MakeConnections_Request

### Community 111 - "internal/cluster (local.go)"
Cohesion: 0.16
Nodes (11): Context, Docker, ListEndpoints_Request, ListEndpoints_Response, ServiceDiscovery, UpsertEndpoints_Request, UpsertEndpoints_Response, Version_Request (+3 more)

### Community 112 - "internal/api (DeleteConfig_Request)"
Cohesion: 0.10
Nodes (17): CreateDeploymentParams, CreateSpecificationParams, GetSpecificationByIdRow, UpdateDeploymentStatusParams, Queries, Tx, Context, Queries (+9 more)

### Community 114 - "Community 114"
Cohesion: 0.18
Nodes (10): CSS, EnvChip, EnvChip, FreezeChip, FreezeChip, IncidentChip, IncidentChip, Notes (+2 more)

### Community 115 - "internal/api (velez_api.pb.go)"
Cohesion: 0.12
Nodes (10): CreateDeploy_Request_Upgrade, CreateSmerd_Request, Context, CreateDeploy_Request, CreateDeploy_Request_Upgrade_, CreateDeploy_Response, Impl, CreateDeploy_Request (+2 more)

### Community 116 - "internal/api (Message)"
Cohesion: 0.05
Nodes (40): Absent, BaseEnableHeadscaleServer, BaseEnablePluginRequest, ConnectSlave, ConnectSlaveRequest, ConnectSlaveResponse, ControlPlaneAPI, CreateEnvironment (+32 more)

### Community 117 - "internal/clients (Filter)"
Cohesion: 0.19
Nodes (5): Args, New(), Filter, health, status

### Community 118 - "internal/cluster (service.go)"
Cohesion: 0.15
Nodes (13): Context, ListSmerds_Request, ListSmerds_Response, TestCreateSmerd_EmptyEnvironmentResolvesDefaultSuffix(), TestCreateSmerd_UnknownEnvironmentRejected(), TestDropSmerd_UnknownEnvironmentRejected(), TestListSmerds_EmptyEnvironmentReachesService(), TestListSmerds_UnknownEnvironmentRejected() (+5 more)

### Community 119 - "internal/api (CreateConfig_Request)"
Cohesion: 0.16
Nodes (12): AboutService, CreateServiceReq, GetServiceReq, RemoveServiceReq, Service, DeploymentStatus, ServiceBaseInfo, Context (+4 more)

### Community 120 - "pkg/web (DeploymentHistory.tsx)"
Cohesion: 0.13
Nodes (16): Custom, ClusterClients, Impl, App, Context, Listener, New(), Context (+8 more)

### Community 124 - ".claude/factory (factory.sh)"
Cohesion: 0.35
Nodes (11): check_deps(), check_ollama(), err(), list_tasks(), log(), main(), ok(), pick_task() (+3 more)

### Community 125 - "Community 125"
Cohesion: 0.16
Nodes (18): App (Velez Dashboard reference mockup), Badge component (mockup), ControlPlaneTab component (mockup), DeployDialog (multi-step deploy modal), DeploymentsTab (kanban + list view), EnvChip component (mockup), FreezChip (release-freeze chip, mockup), IncidentChip component (mockup) (+10 more)

### Community 127 - "internal/api (ConfigBase)"
Cohesion: 0.25
Nodes (8): Context, Mutex, Queries, Tx, VelezJob, jobKey(), newJobsStorage(), jobs

### Community 129 - "tests/e2e (TestEnvironment)"
Cohesion: 0.17
Nodes (9): commonRuntime, labelBasedRuntime, APIClient, Context, CreateResponse, InspectResponse, ListSmerds_Request, Summary (+1 more)

### Community 131 - "internal/transport (.CreateDeploy())"
Cohesion: 0.18
Nodes (9): Configurator, Container, ContainerManager, Context, Docker, Service, VervService, New() (+1 more)

### Community 132 - "internal/api (MessageState)"
Cohesion: 0.09
Nodes (7): MessageState, SizeCache, UnknownFields, Connection, Container, FileConfig, PlainConfigSpec

### Community 133 - "internal/pipelines (Step)"
Cohesion: 0.27
Nodes (7): createContainerStep, Create(), APIClient, Context, CreateRequest, Docker, Step

### Community 135 - "internal/storage (services.go)"
Cohesion: 0.23
Nodes (11): containerStateToDeploymentStatus(), containerStateToString(), countRunningServices(), Context, DeploymentStatus, Docker, Service, ServiceBaseInfo (+3 more)

### Community 136 - "pkg/web (ResourcesSection.tsx)"
Cohesion: 0.10
Nodes (11): Pointer, NewContainer(), APIClient, ContainerState, Context, EndpointSettings, InspectResponse, PauseContainer() (+3 more)

### Community 140 - "Community 140"
Cohesion: 0.22
Nodes (6): Closable, Context, Client, Context, Client, CloseWithLog()

### Community 142 - "internal/pipelines (create.go)"
Cohesion: 0.08
Nodes (4): isPatch_Patch, Patch, PatchConfig_Request, SubscribeOnChanges_Response

### Community 143 - "internal/api (CreateDeploy_Request)"
Cohesion: 0.09
Nodes (21): Connect(), ConnectToContainer(), getAPIAddress(), Context, Docker, Client, Container, Docker (+13 more)

### Community 146 - "internal/api (MessageState)"
Cohesion: 0.18
Nodes (3): SizeCache, MakeConnections, MakeConnections_Response

### Community 147 - "internal/api (SizeCache)"
Cohesion: 0.15
Nodes (13): dependencies, classnames, framer-motion, @microlink/react-json-view, react, react-dom, react-router-dom, react-tooltip (+5 more)

### Community 148 - "internal/api (UnknownFields)"
Cohesion: 0.18
Nodes (17): Context, resolveEnvironmentSuffix(), newEnvProvider(), newEnvStorage(), TestCreateContainerJob_EmptyEnvironmentStampsDefaultSuffix(), TestCreateContainerJob_StampsResolvedEnvironmentSuffix(), TestCreateContainerJob_UnknownEnvironmentFailsJob(), TestCreateSmerdTaskPayload_EnvironmentSurvivesRoundTrip() (+9 more)

### Community 151 - "internal/transport (network.go)"
Cohesion: 0.22
Nodes (8): BreakConnections_Request, BreakConnections_Response, Connection, Context, MakeConnections_Request, MakeConnections_Response, Impl, toConnection()

### Community 152 - "pkg/web (ControlPlanePage.tsx)"
Cohesion: 0.07
Nodes (49): EnableStatefullTaskPayload, App, Checkpoint(), Context, TestCheckpoint_FailFastWhenFailed(), TestCheckpoint_RunsAndPersistsContextOnSuccess(), TestCheckpoint_SkipsWhenDone(), NewCreateServiceHandler() (+41 more)

### Community 153 - "internal/api (EnableStatefullCluster)"
Cohesion: 0.22
Nodes (8): Deployment, DeploymentList, DeploymentSpecification, ListDeploymentsReq, Paging, Time, VelezDeploymentStatus, Context

### Community 156 - "internal/transport (VcnNamespace)"
Cohesion: 0.08
Nodes (22): New(), NewTxManager(), NewPg(), Queries, newDeploymentsStorage(), Queries, newJobsStorage(), Queries (+14 more)

### Community 159 - "tests/e2e (suite_hello_world_cluster_test.go)"
Cohesion: 0.26
Nodes (8): Cors, Handler, Listener, ServeMux, Server, newHttpServer(), setUpCors(), httpServer

### Community 161 - "pkg/web (VcnApi)"
Cohesion: 0.20
Nodes (13): FromPorts(), Container_Settings, ToPort(), ToPortProtocol(), ToPortsMapping(), ToPortsSlice(), ContainerManager, Context (+5 more)

### Community 162 - "pkg/web (compilerOptions)"
Cohesion: 0.22
Nodes (8): compilerOptions, allowSyntheticDefaultImports, composite, module, moduleResolution, skipLibCheck, strict, include

### Community 165 - "internal/cluster (Synchronizer)"
Cohesion: 0.25
Nodes (5): Synchronizer, Client, Context, MatreshkaBeAPI_SubscribeOnChangesClient, New()

### Community 167 - "internal/domain (images.go)"
Cohesion: 0.31
Nodes (7): ImageListRequest, ImageSearchRequest, APIClient, Context, Image, ListImages(), SearchImages()

### Community 168 - "internal/cluster (DisabledVcnImpl)"
Cohesion: 0.11
Nodes (13): AppConfig, ConfigMeta, ConfigFormat, ConfigTypePrefix, Configurator, Context, Configurator, Context (+5 more)

### Community 170 - "internal/clients (api_key_issue.go)"
Cohesion: 0.29
Nodes (8): New(), Context, VelezTask, Test_CreateSmerdStream(), Test_CreateSmerdStream_EnqueueError(), enqueueCall, fakeJobsEngine, watchCall

### Community 171 - "internal/api (matreshka_common.pb.go)"
Cohesion: 0.11
Nodes (15): Bool, CMux, Context, Listener, ServeMux, ServerOption, newGrpcServer(), Context (+7 more)

### Community 173 - "internal/api (verv_closed_network.pb.go)"
Cohesion: 0.25
Nodes (3): file_verv_closed_network_proto_init(), init(), ConnectUser

### Community 174 - "internal/storage (service_dependencies.sql.go)"
Cohesion: 0.33
Nodes (5): Context, Queries, GetServiceCallersRow, GetServiceDependenciesRow, UpsertServiceDependencyParams

### Community 177 - "internal/api (Container_Hardware)"
Cohesion: 0.12
Nodes (16): APIClient, Context, copyToContainerJob, createContainerJob, createSmerdImageAccessor, dropOwnedContainerJob, healthcheckJob, imageExposedPortsAccessor (+8 more)

### Community 180 - "internal/api (Image)"
Cohesion: 0.15
Nodes (9): Docker, FromContainerToRequest(), Context, NetworkBind, Smerd, prepareUpgradeImageJob, upgradeSmerdHandler, ContainerService (+1 more)

### Community 183 - "internal/api (SearchImageItem)"
Cohesion: 0.19
Nodes (9): fakeListContainerRuntime, ContainerRuntime, Context, Context, ListSmerds_Request, Summary, Context, Environment (+1 more)

### Community 186 - "internal/pipelines (get_root_dsn.go)"
Cohesion: 0.16
Nodes (10): VelezJob, Context, Queries, NullString, VelezJob, VelezJobStatus, fakeJobsStorage, CreateRunningJobParams (+2 more)

### Community 188 - "internal/domain (vcn.go)"
Cohesion: 0.22
Nodes (8): Backlog (suggested order — confirm before starting if you'd rather reorder), Context, Ground rules (apply to every task below, every agent), Open decisions (need explicit user sign-off before the affected work starts), Per-pipeline recipe (repeat for each row in the backlog), Pipeline #1 special case: `CreateSmerd` streaming pilot, Plan: Cut Pipelines Over to the Postgres/Storage-Based Jobs Engine, Progress log

### Community 189 - "internal/patterns (headscale.go)"
Cohesion: 0.27
Nodes (6): Context, Once, Ticker, deploymentsList, deployWatcher, environmentsProvider

### Community 193 - "internal/patterns (.EnableStatefullMode())"
Cohesion: 0.11
Nodes (18): ListServicesReq, ServiceBaseInfo, ServiceList, Paging, Time, Context, ServiceBaseInfo, VervService (+10 more)

### Community 194 - "fetchSmerdConfigJob"
Cohesion: 0.20
Nodes (11): fakeInspectOnlyAPI, fakeListAPI, fakeRemoveAPI, fakeRenameAPI, renameCall, APIClient, Context, InspectResponse (+3 more)

### Community 195 - "internal/api (GetConfigNode_Request)"
Cohesion: 0.20
Nodes (9): Artel "Tract", Common ground (both systems converged independently — kept in the merge), Divergent decisions — resolved, Open trade-off to confirm before implementation, Per-system deep dive, Sources, Synthesized target architecture, Task/Job Engine — Velez vs. Artel Comparison & Merge Direction (+1 more)

### Community 196 - "internal/api (GetConfigNode_Response)"
Cohesion: 0.09
Nodes (20): Agent code of conduct, API (Proto definitions in `api/grpc/`), Architecture, Building, Code Generation, Code Style, Commands, Configuration (+12 more)

### Community 197 - "UpsertPluginParams"
Cohesion: 0.23
Nodes (8): Context, NodeBaseInfo, Querier, SelectBuilder, newNodeStorage(), scanNode(), nodeStorage, Scannable

### Community 208 - ".claude/factory (review.sh)"
Cohesion: 0.52
Nodes (5): err(), log(), ok(), review_task(), review.sh script

### Community 209 - "internal/storage (db.go)"
Cohesion: 0.38
Nodes (5): DBTX, Queries, Queries, Tx, New()

### Community 212 - "Version_Response"
Cohesion: 0.40
Nodes (3): fromApiNodes(), Configurator, Context

### Community 213 - "internal/clients (.doApiRequest())"
Cohesion: 0.43
Nodes (4): Context, Client, Request, Response

### Community 214 - "internal/api (file_matreshka_common_proto_rawDescGZIP())"
Cohesion: 0.33
Nodes (5): CheckSidecarExist(), Context, Docker, Step, checkSidecarExist

### Community 216 - "T08 — SectionLabel"
Cohesion: 0.33
Nodes (5): Component, CSS, Props interface, T16 — Sidebar Widget, What it looks like

### Community 217 - "internal/storage (db.go)"
Cohesion: 0.38
Nodes (5): Queries, Queries, Tx, New(), DBTX

### Community 218 - "internal/storage (db.go)"
Cohesion: 0.38
Nodes (5): Queries, Queries, Tx, New(), DBTX

### Community 219 - "internal/storage (db.go)"
Cohesion: 0.38
Nodes (5): Queries, Queries, Tx, New(), DBTX

### Community 220 - "internal/storage (db.go)"
Cohesion: 0.38
Nodes (5): Queries, Queries, Tx, New(), DBTX

### Community 221 - "internal/storage (service_resources.sql.go)"
Cohesion: 0.38
Nodes (4): Context, Queries, GetServiceResourcesRow, UpsertServiceResourceParams

### Community 222 - "internal/storage (db.go)"
Cohesion: 0.38
Nodes (5): Queries, Queries, Tx, New(), DBTX

### Community 223 - "pkg/web (ErrorBoundary.tsx)"
Cohesion: 0.22
Nodes (9): scripts, build, dev, gen, knip, lint, lint:css, lint:js (+1 more)

### Community 225 - "internal/api (CreateService_Request)"
Cohesion: 0.33
Nodes (5): Component, CSS, Props interface, T15 — CodeBlock, What it looks like

### Community 228 - "internal/api (GetService_Request)"
Cohesion: 0.24
Nodes (8): Context, ListOptions, ListSmerds_Request, Summary, Volume, ListResponse, fakeCascadeDocker, fakeCascadeNetVol

### Community 230 - "internal/api (GetServiceMetrics_Request)"
Cohesion: 0.16
Nodes (12): NodeType, ServiceDependency, ServiceEnvironment, ServiceGraph, Time, Context, VervService, containerName() (+4 more)

### Community 231 - "internal/api (GetServiceResources_Request)"
Cohesion: 0.10
Nodes (17): HealthConfig, FromCommand(), FromHealthcheck(), Container_Healthcheck, FromDockerEnv(), ToDockerEnv(), FromRestart(), RestartPolicy (+9 more)

### Community 232 - "Message"
Cohesion: 0.17
Nodes (9): Context, Handler, ServiceRegistrar, Impl, New(), ConnectService_Request, ConnectService_Response, Context (+1 more)

### Community 234 - "internal/api (StopService_Request)"
Cohesion: 0.25
Nodes (3): file_velez_api_proto_init(), init(), BreakConnections

### Community 236 - "Volume"
Cohesion: 0.26
Nodes (6): Context, Duration, Once, Ticker, VelezTask, taskWorker

### Community 237 - "tests/e2e (suite_control_plane_test.go)"
Cohesion: 0.13
Nodes (5): EnumDescriptor, EnumNumber, EnumType, TaskStatus_JobStatus, TaskStatus_Status

### Community 238 - "internal/clients (list_containers.go)"
Cohesion: 0.33
Nodes (5): APIClient, Context, ListSmerds_Request, Summary, ListContainers()

### Community 239 - "internal/clients (Filter)"
Cohesion: 0.47
Nodes (3): Args, New(), Filter

### Community 240 - "Community 240"
Cohesion: 0.13
Nodes (13): API calls, Architecture, Coding Rules, Commands, Environment, Exploration Rules, Layer structure, Proto regeneration (+5 more)

### Community 241 - "selectiveFailDocker"
Cohesion: 0.11
Nodes (17): ListNodesReq, NodeBaseInfo, NodesList, NodeBaseInfo, Paging, Time, Context, Storage (+9 more)

### Community 242 - "pkg/web (.GetService())"
Cohesion: 0.40
Nodes (4): Context, RemoveService_Request, RemoveService_Response, Impl

### Community 243 - "pkg/web (MockDeployRow.tsx)"
Cohesion: 0.12
Nodes (3): Message, ConnectSlave, ListVervPlugins_Request

### Community 244 - "NewServer"
Cohesion: 0.27
Nodes (3): TaskStatus, MD, fakeTaskStatusStream

### Community 245 - "TasksStorage"
Cohesion: 0.14
Nodes (4): file_velez_common_proto_init(), init(), Port, Port_Protocol

### Community 248 - "internal/api (GetHardware_Request)"
Cohesion: 0.31
Nodes (6): getConfigFromContainerStep, GetConfigFromContainerStep(), APIClient, Context, InspectResponse, MatreshkaConfigSpec

### Community 249 - "internal/api (InitMaster_Response)"
Cohesion: 0.08
Nodes (8): storeConfigStep, Configurator, Context, Step, StoreConfig(), Format, GetConfig_Request, SaveConfig_Request

### Community 250 - "internal/api (ListDeployments)"
Cohesion: 0.29
Nodes (6): Context, ServiceBaseInfo, Step, Storage, UpsertServiceState(), upsertServiceState

### Community 251 - "internal/api (ListEnvironments)"
Cohesion: 0.22
Nodes (3): DropSmerd_Request, DropSmerd_Response_Error, DropSmerdTaskPayload

### Community 253 - "Exec"
Cohesion: 0.21
Nodes (14): TestCreateNewDeploy_GetByNameError(), TestUpgradeDeploy_GetByNameError(), cascadeRemoveEnvironmentResources(), Context, VervService, removeEnvironmentContainers(), removeEnvironmentNetworks(), removeEnvironmentVolumes() (+6 more)

### Community 257 - "GetServiceMetrics"
Cohesion: 0.26
Nodes (5): FileConfig, MatreshkaConfigSpec, ConfigTypePrefix, toConfTypePrefix(), fetchSmerdConfigJob

### Community 260 - "ListNodes"
Cohesion: 0.33
Nodes (5): entry, ignore, project, $schema, knip

### Community 261 - "internal/api (UpgradeSmerd)"
Cohesion: 0.18
Nodes (3): UnknownFields, AssembleConfig, UpgradeSmerd

### Community 262 - "internal/api (UpgradeSmerd_Response)"
Cohesion: 0.12
Nodes (11): Context, Step, ValidateServiceName(), Context, SingleFunc(), createServiceHandler, serviceNameAccessor, upsertServiceJob (+3 more)

### Community 263 - "createPgUserStep"
Cohesion: 0.22
Nodes (8): Investigation: `Test_Lifecycle` flakiness (Test_ClusterMode_* subtests), Open decision (do not implement without sign-off — see repo `CLAUDE.md`:, Recommended next step, Reproduction, Resolution, Root cause, Three more pre-existing bugs found while verifying — two fixed, one still open, What was ruled out

### Community 264 - ".claude/factory (new-task.sh)"
Cohesion: 0.70
Nodes (4): err(), log(), ok(), new-task.sh script

### Community 266 - "internal/service (.ConnectToNetwork())"
Cohesion: 0.60
Nodes (3): Connection, ContainerManager, Context

### Community 267 - "internal/service (.DropSmerds())"
Cohesion: 0.40
Nodes (4): ContainerManager, Context, DropSmerd_Request, DropSmerd_Response

### Community 268 - "internal/transport (.ListEnvironments())"
Cohesion: 0.29
Nodes (10): fakeInspectAPI, APIClient, ContainerManager, Context, InspectResponse, newInspectContainerResponse(), newTestContainerManager(), TestInspectSmerd_EmptySuffixLeavesNameUntouched() (+2 more)

### Community 269 - "internal/transport (.ListPlugins())"
Cohesion: 0.40
Nodes (4): Context, Impl, ListPlugins_Request, ListPlugins_Response

### Community 271 - "internal/clients (hardware_manager.go)"
Cohesion: 0.20
Nodes (10): Manager, Context, GetHardware_Response, Mutex, Time, New(), TestManager_GetHardware_CachesWithinTTL(), TestManager_GetHardware_IsRunningInContainer() (+2 more)

### Community 275 - "internal/clients (pull_image.go)"
Cohesion: 0.40
Nodes (4): APIClient, Context, Image, PullImage()

### Community 276 - "InitMaster_Response"
Cohesion: 0.18
Nodes (3): MessageState, BreakConnections_Response, CreateSmerd

### Community 277 - "VelezAPI"
Cohesion: 0.15
Nodes (14): Context, CreateSmerd_Request, ServerStream, ServerStreamingServer, ServiceRegistrar, TaskStatus, WatchTask_Request, RegisterTasksApiServer() (+6 more)

### Community 278 - "internal/storage (.GetByName())"
Cohesion: 0.47
Nodes (3): Context, Queries, VelezService

### Community 279 - "internal/transport (.GetServiceGraph())"
Cohesion: 0.40
Nodes (4): Context, GetServiceGraph_Request, GetServiceGraph_Response, Impl

### Community 280 - "internal/transport (.GetServiceMetrics())"
Cohesion: 0.40
Nodes (4): Context, GetServiceMetrics_Request, GetServiceMetrics_Response, Impl

### Community 281 - "internal/transport (.GetServiceResources())"
Cohesion: 0.40
Nodes (4): Context, GetServiceResources_Request, GetServiceResources_Response, Impl

### Community 282 - "localStorage"
Cohesion: 0.32
Nodes (5): CreateEnvironmentParams, UpdateEnvironmentParams, Context, Queries, VelezEnvironment

### Community 283 - "internal/transport (.GetService())"
Cohesion: 0.40
Nodes (4): Context, GetService_Request, GetService_Response, Impl

### Community 284 - "internal/transport (.RestartService())"
Cohesion: 0.40
Nodes (4): Context, RestartService_Request, RestartService_Response, Impl

### Community 285 - "internal/transport (.StopService())"
Cohesion: 0.40
Nodes (4): Context, Impl, StopService_Request, StopService_Response

### Community 286 - "internal/transport (.GetVervonomicon())"
Cohesion: 0.40
Nodes (4): Context, GetVervonomicon_Request, GetVervonomicon_Response, Impl

### Community 287 - "internal/transport (.ListPeers())"
Cohesion: 0.40
Nodes (4): Context, ListPeers_Request, ListPeers_Response, Impl

### Community 288 - "internal/transport (.DeleteNamespace())"
Cohesion: 0.40
Nodes (4): Context, DeleteVcnNamespace_Request, DeleteVcnNamespace_Response, Impl

### Community 290 - "portManagerImpl"
Cohesion: 0.47
Nodes (3): UpgradeDeployReq, Context, VervService

### Community 291 - "internal/transport (.ConnectUser())"
Cohesion: 0.40
Nodes (4): ConnectUser_Request, ConnectUser_Response, Context, Impl

### Community 292 - "internal/transport (.AssembleConfig())"
Cohesion: 0.40
Nodes (4): AssembleConfig_Request, AssembleConfig_Response, Context, Impl

### Community 293 - "internal/transport (.GetHardware())"
Cohesion: 0.40
Nodes (4): Context, GetHardware_Request, GetHardware_Response, Impl

### Community 294 - "internal/transport (.Version())"
Cohesion: 0.40
Nodes (4): Context, Impl, Version_Request, Version_Response

### Community 295 - "internal/transport (.SearchImages())"
Cohesion: 0.40
Nodes (4): Context, SearchImages_Request, SearchImages_Response, Impl

### Community 296 - "internal/transport (.CreateSmerd())"
Cohesion: 0.40
Nodes (4): Context, CreateSmerd_Request, Smerd, Impl

### Community 297 - "internal/transport (.DropSmerd())"
Cohesion: 0.40
Nodes (4): Context, DropSmerd_Request, DropSmerd_Response, Impl

### Community 298 - "internal/transport (.ListSmerds())"
Cohesion: 0.40
Nodes (4): Context, ListSmerds_Request, ListSmerds_Response, Impl

### Community 299 - "internal/transport (.UpgradeSmerd())"
Cohesion: 0.40
Nodes (4): Context, UpgradeSmerd_Request, UpgradeSmerd_Response, Impl

### Community 300 - "schemas/Smerds.md (Matreshka (external config service))"
Cohesion: 1.00
Nodes (3): Matreshka (external config service), hello_world.yaml Matreshka Config Mock, velez_default_config.yaml Matreshka Config Mock

### Community 301 - "internal/clients (matreshka.go)"
Cohesion: 0.67
Nodes (3): MatreshkaBeAPIClient, NewClient(), Client

### Community 302 - "internal/service (.GetServiceEnvironments())"
Cohesion: 0.50
Nodes (3): Context, VervService, ServiceEnvironment

### Community 303 - "internal/service (.GetServiceMetrics())"
Cohesion: 0.50
Nodes (3): Context, VervService, ServiceMetrics

### Community 304 - "internal/service (.GetServiceResources())"
Cohesion: 0.50
Nodes (3): BoundResource, Context, VervService

### Community 306 - "internal/storage (Context)"
Cohesion: 0.32
Nodes (4): Context, Queries, NullFloat64, UpdateOnlineParams

### Community 307 - "start/run_velez.sh (run_velez.sh)"
Cohesion: 0.50
Nodes (3): IMAGE, run_velez.sh script, VELEZ_PORT_GRPC

### Community 308 - "pluginsStorage"
Cohesion: 0.20
Nodes (13): extract_message(), main(), WithExposedPort(), WithInstanceName(), WithPassword(), WithPort(), basicPostgresConstructor(), CreateRequest (+5 more)

### Community 310 - "PostgreSQL Database"
Cohesion: 0.67
Nodes (3): PostgreSQL Database, Database Pixel Icon, PostgreSQL Service Icon

### Community 311 - "CreatePgUserForNode"
Cohesion: 0.40
Nodes (10): environmentColumns(), newPgTest(), TestPgStorage_CreateEnvironment(), TestPgStorage_DeleteEnvironment(), TestPgStorage_GetEnvironmentByID(), TestPgStorage_GetEnvironmentByName(), TestPgStorage_GetEnvironmentByName_NotFound(), TestPgStorage_ListEnvironments() (+2 more)

### Community 317 - "internal/service (.ListEnvironments())"
Cohesion: 0.43
Nodes (5): Context, Docker, InspectResponse, PrepareImage(), stepPrepareImage

### Community 319 - "Community 319"
Cohesion: 0.67
Nodes (3): Arrow Forward Icon, Open In Tab Icon, Rocket Icon

### Community 321 - "Input.tsx"
Cohesion: 0.31
Nodes (6): copyToContainerStep, ConfigMount, FileMountPoint, CopyToContainer(), APIClient, Context

### Community 322 - "ContainerCreateRequest"
Cohesion: 0.21
Nodes (12): ContainerConfig, ContainerCreateRequest, fakeCreateAPI, HostConfig, NetworkingConfig, CreateResponse, Platform, Platform (+4 more)

### Community 323 - "pkg/web (TextInput.tsx)"
Cohesion: 0.40
Nodes (4): name, private, trustedDependencies, type

### Community 325 - "selectiveFailDocker"
Cohesion: 0.27
Nodes (6): calculatePluginState(), Context, Querier, VervPlugin_State, newPluginsStorage(), pluginsStorage

### Community 340 - "pkg/web (eslint)"
Cohesion: 0.13
Nodes (13): Velez — Features, `auth.yaml` — inter-service access control, `configuration.yaml` — app configuration, `dependencies.yaml` — external resources, Directory layout, How Velez processes the descriptor, `network.yaml` — ports and exposure, `resources.yaml` — hardware and placement (+5 more)

### Community 342 - "runner"
Cohesion: 0.31
Nodes (5): CreateDeployReq, Context, fakeDeployService, fakeEnvResolver, VervServicesService

### Community 351 - "pkg/web (react)"
Cohesion: 0.13
Nodes (14): Access key, Advanced, Basic v1, Default port, Docker run script, Download and start, Features checklist, generated with love for coding by [RedSock CLI](https://github.com/Red-Sock/rscli) (+6 more)

### Community 352 - "pkg/web (react-highlight)"
Cohesion: 0.22
Nodes (18): ClientConnInterface, NewTasksApiClient(), ClientConn, Context, DialOption, Marshaler, Request, ServeMux (+10 more)

### Community 353 - "pkg/web (@tanstack/react-query)"
Cohesion: 0.20
Nodes (6): M3 — Observability, Scope (draft), M4 — Access & Multi-tenancy, Scope (draft), M5 — PaaS Automation, Scope (draft)

### Community 354 - "pkg/web (vite)"
Cohesion: 0.33
Nodes (5): Component, CSS, Props interface, T06 — IconButton, What it looks like

### Community 361 - "ConnectSlave_Response"
Cohesion: 0.17
Nodes (8): subscribeForConfigChangesStep, unsubscribeForConfigChangesStep, Context, SubscribeForConfigChanges(), Context, UnSubscribeForConfigChanges(), createSmerdHandler, ConfigurationService

### Community 365 - "Impl"
Cohesion: 0.23
Nodes (8): Context, CreateSmerd_Request, Handler, ServerStreamingServer, ServiceRegistrar, TaskStatus, WatchTask_Request, Impl

### Community 367 - ".enrichServiceWithSmerdData"
Cohesion: 0.15
Nodes (13): Connection, Context, DropSmerd_Request, DropSmerd_Response, ListSmerds_Request, ListSmerds_Response, Service, Smerd (+5 more)

### Community 368 - "internal/clients (ns_list.go)"
Cohesion: 0.28
Nodes (6): getPgDbDsn, GetRgRootDsn(), Context, Docker, InspectResponse, Step

### Community 372 - ".GetServiceEnvironments"
Cohesion: 0.31
Nodes (6): MakoshBeAPIClient, NewClient(), NewServiceDiscovery(), HeaderOutgoingInterceptor(), ServiceDiscovery, UnaryClientInterceptor

### Community 387 - "Container_Hardware"
Cohesion: 0.28
Nodes (6): Context, Storage, NewPgStateManager(), Context, SetupPgState(), pgState

### Community 429 - "Velez UI Redesign — Roadmap"
Cohesion: 0.22
Nodes (8): 1. Start the first node and let it self-host Headscale, 2. Point a second node at node A's Headscale server (manual — no join RPC exists), 3. Create a VCN namespace and attach a service to it (this part genuinely works), 4. Deploying a service "across the cluster" — current limits, Prerequisites, Quickstart: Multi-Node, Troubleshooting / gotchas, What's real today vs. what's a stub

### Community 431 - "T30 — AppsPage + AppCard (logical services view)"
Cohesion: 0.17
Nodes (11): AppCard CSS highlights, AppCard data type, AppCard layout, AppCard props, AppsPage component, AppsPage CSS, Notes, Routes update (+3 more)

### Community 432 - "EnableStatefullCluster"
Cohesion: 0.33
Nodes (5): AddMakoshRecord(), Context, ServiceDiscovery, Step, addMakoshRecord

### Community 434 - "Tasks"
Cohesion: 0.20
Nodes (9): 5.1 Loading states, 5.2 Error boundaries, 5.3 Empty states, 5.4 Toast / notification system, 5.5 Navigation & layout cleanup, Acceptance criteria, Goal, T5 — UX Polish (+1 more)

### Community 435 - "FileMountPoint"
Cohesion: 0.22
Nodes (8): 1. Start Velez, 2. Find your access key, 3. Create a service, 4. Create a deployment, 5. Verify it's running, Prerequisites, Quickstart: Single Node, Troubleshooting / gotchas

### Community 437 - "GetConfigNode_Response"
Cohesion: 0.12
Nodes (6): MessageState, SizeCache, UnknownFields, CopyToVolumeTaskPayload, CreateServiceTaskPayload, WatchTask

### Community 438 - "B1-T01 — Extend API for UI wiring (backend)"
Cohesion: 0.22
Nodes (8): 1. Enrich `ServiceBaseInfo` in `ListServices`, 2. Add `ListNodes` endpoint, 3. Add VCN peer details to `ListNamespaces`, Acceptance Criteria, B1-T01 — Extend API for UI wiring (backend), Files to create / modify, Goal, Notes

### Community 439 - "Task B2-T01 — Service Runtime Stats API"
Cohesion: 0.22
Nodes (8): Acceptance Criteria, Context, Do NOT change, Files to Create / Modify, Goal, Notes, Proto Changes, Task B2-T01 — Service Runtime Stats API

### Community 440 - "Tasks"
Cohesion: 0.14
Nodes (12): Approach, Exit criteria, M1 — Core Platform, Task groups, 1.1 Services list (HomePage), 1.2 Service detail (ServiceInfoPage), 1.3 Smerds list (smerd containers), 1.4 Smerd detail (SmerdPage) (+4 more)

### Community 441 - "Tasks"
Cohesion: 0.22
Nodes (8): 2.1 Deployments list (inside ServiceInfoPage), 2.2 Deployment detail drawer / page, 2.3 Deploy flow (DeployPage / DeployMenu), 2.4 "Deploy latest" shortcut, Acceptance criteria, Goal, T2 — Deployments, Tasks

### Community 442 - "Tasks"
Cohesion: 0.22
Nodes (8): 3.1 Verv services list (HomePage or dedicated page), 3.2 New service form (NewServicePage), 3.3 Service delete, 3.4 Edit service (stretch goal for M1), Acceptance criteria, Goal, T3 — Verv Services Management, Tasks

### Community 443 - "Tasks"
Cohesion: 0.22
Nodes (8): 4.1 Settings widget (already partially exists), 4.2 Connection health indicator, 4.3 Settings validation, 4.4 Environment variable display (informational), Acceptance criteria, Goal, T4 — Settings Panel, Tasks

### Community 444 - "Container_Healthcheck"
Cohesion: 0.29
Nodes (6): SetupHeadscaleRequest, SetupHeadscaleResponse, BasicConfig(), CreateRequest, Headscale(), pipeliner

### Community 445 - "WatchTask_Request"
Cohesion: 0.22
Nodes (3): TaskStatus_JobStatus, Timestamp, TaskStatus

### Community 447 - "ListPlugins_Request"
Cohesion: 0.47
Nodes (5): EnvironmentsProvider, APIClient, ContainerManager, Docker, New()

### Community 448 - "Project Factory — Agent Context"
Cohesion: 0.25
Nodes (7): Agent Rules, Architecture, Code Style, Frontend, Go, Project Factory — Agent Context, Task System

### Community 449 - "Task 000 — Example Task Title"
Cohesion: 0.25
Nodes (7): Acceptance Criteria, Context, Do NOT change, Files to Create / Modify, Goal, Notes, Task 000 — Example Task Title

### Community 450 - "Task 001 — Health check endpoint"
Cohesion: 0.25
Nodes (7): Acceptance Criteria, Context, Do NOT change, Files to Create / Modify, Goal, Notes, Task 001 — Health check endpoint

### Community 452 - "Task B3-T01 — Plugin Service with Dual-Mode Storage and Hot-Switch"
Cohesion: 0.25
Nodes (7): Acceptance Criteria, Context, Do NOT change, Files to Create / Modify, Goal, Notes, Task B3-T01 — Plugin Service with Dual-Mode Storage and Hot-Switch

### Community 453 - "M2 — Cluster & Networking"
Cohesion: 0.25
Nodes (7): Dependencies, Goal, M2 — Cluster & Networking, Scope (draft), T-node-tags — Node Scheduling Tags, Tags, Tasks

### Community 454 - "Task 039 — PluginManageDialog: enable action with per-plugin config forms"
Cohesion: 0.25
Nodes (7): Acceptance Criteria, Context, Do NOT change, Files to Create / Modify, Goal, Notes, Task 039 — PluginManageDialog: enable action with per-plugin config forms

### Community 455 - "T22 — PluginMatrix Widget"
Cohesion: 0.25
Nodes (7): Component, CSS, Data types, Notes, Props interface, T22 — PluginMatrix Widget, What it looks like

### Community 456 - "T23 — NetworkTopologyMap Widget"
Cohesion: 0.25
Nodes (7): Component, CSS, Data types, Notes, Props interface, T23 — NetworkTopologyMap Widget, What it looks like

### Community 457 - "T25 — MainLayout (rebuild)"
Cohesion: 0.25
Nodes (7): API data, Component, Current layout (to replace), New layout structure, Notes, T25 — MainLayout (rebuild), What it does

### Community 458 - "T29 — SearchPage"
Cohesion: 0.25
Nodes (7): Component, CSS, Notes, Props, Routing, T29 — SearchPage, What it looks like

### Community 459 - "Task 032 — PluginMatrix: enable/disable actions and service page navigation"
Cohesion: 0.25
Nodes (7): Acceptance Criteria, Context, Do NOT change, Files to Create / Modify, Goal, Notes, Task 032 — PluginMatrix: enable/disable actions and service page navigation

### Community 460 - "Task M9-T33 — ServiceInfoPage: Full Redesign"
Cohesion: 0.25
Nodes (7): Acceptance Criteria, Context, Do NOT change, Files to Create / Modify, Goal, Notes, Task M9-T33 — ServiceInfoPage: Full Redesign

### Community 461 - "Task M9-T34 — ObservabilityLinksPanel Widget"
Cohesion: 0.25
Nodes (7): Acceptance Criteria, Context, Do NOT change, Files to Create / Modify, Goal, Notes, Task M9-T34 — ObservabilityLinksPanel Widget

### Community 462 - "Task M9-T35 — Environment Tab Switcher and Tags Strip"
Cohesion: 0.25
Nodes (7): Acceptance Criteria, Context, Do NOT change, Files to Create / Modify, Goal, Notes, Task M9-T35 — Environment Tab Switcher and Tags Strip

### Community 463 - "Task M9-T36 — Sidebar: Fix Active Nav Highlight"
Cohesion: 0.25
Nodes (7): Acceptance Criteria, Context, Do NOT change, Files to Create / Modify, Goal, Notes, Task M9-T36 — Sidebar: Fix Active Nav Highlight

### Community 464 - "Task M9-T37 — Settings Page"
Cohesion: 0.25
Nodes (7): Acceptance Criteria, Context, Do NOT change, Files to Create / Modify, Goal, Notes, Task M9-T37 — Settings Page

### Community 465 - "Task 038 — PluginMatrix: restore status display + separate Manage button with empty dialog"
Cohesion: 0.25
Nodes (7): Acceptance Criteria, Context, Do NOT change, Files to Create / Modify, Goal, Notes, Task 038 — PluginMatrix: restore status display + separate Manage button with empty dialog

### Community 466 - "GetVervonomicon"
Cohesion: 0.38
Nodes (5): DBTX, Queries, Queries, Tx, New()

### Community 467 - "VervServicesService"
Cohesion: 0.36
Nodes (5): BoundResource, Context, Docker, newServiceResourcesStorage(), dockerServiceResourcesStorage

### Community 468 - "🏭 Coding Factory"
Cohesion: 0.29
Nodes (6): 🏭 Coding Factory, Cost, Daily Workflow, Pipeline Architecture, Setup, Task File Format

### Community 469 - "Local state"
Cohesion: 0.29
Nodes (6): Configuration, Headscale, Local state, MatreshkaKey, Network, VelezKey

### Community 470 - "T02 — StatusDot"
Cohesion: 0.29
Nodes (6): Component, CSS, Notes, Props interface, T02 — StatusDot, What it looks like

### Community 471 - "T03 — Badge"
Cohesion: 0.29
Nodes (6): Component, CSS, Notes, Props interface, T03 — Badge, What it looks like

### Community 472 - "T04 — MiniBar (progress bar)"
Cohesion: 0.29
Nodes (6): Component, CSS, Notes, Props interface, T04 — MiniBar (progress bar), What it looks like

### Community 473 - "T07 — Button (rebuild)"
Cohesion: 0.29
Nodes (6): Component, CSS, Notes, Props interface, T07 — Button (rebuild), What it looks like

### Community 474 - "T09 — StatCard"
Cohesion: 0.29
Nodes (6): Component, CSS, Notes, Props interface, T09 — StatCard, What it looks like

### Community 475 - "T10 — ThreeDotMenu"
Cohesion: 0.29
Nodes (6): Component, CSS, Notes, Props interface, T10 — ThreeDotMenu, What it looks like

### Community 476 - "T11 — ServiceCard (Kanban card)"
Cohesion: 0.29
Nodes (6): Component, CSS, Data types, Props interface, T11 — ServiceCard (Kanban card), What it looks like

### Community 477 - "T12 — ServiceListRow (table row)"
Cohesion: 0.29
Nodes (6): Component, CSS, Data types (same as ServiceCard, import from there), Props interface, T12 — ServiceListRow (table row), What it looks like

### Community 478 - "T13 — NodeCard"
Cohesion: 0.29
Nodes (6): Component, CSS, Data types, Props interface, T13 — NodeCard, What it looks like

### Community 479 - "T14 — VCNPeerRow"
Cohesion: 0.29
Nodes (6): Component, CSS, Data types, Props interface, T14 — VCNPeerRow, What it looks like

### Community 480 - "T18 — DeploymentFilters Widget (toolbar)"
Cohesion: 0.29
Nodes (6): Component, CSS, Notes, Props interface, T18 — DeploymentFilters Widget (toolbar), What it looks like

### Community 481 - "T27 — DeploymentsPage (rebuild)"
Cohesion: 0.29
Nodes (6): Component, CSS, Mock data, Notes, T27 — DeploymentsPage (rebuild), What it does

### Community 483 - "db.go"
Cohesion: 0.38
Nodes (5): Queries, Queries, Tx, New(), DBTX

### Community 484 - "db.go"
Cohesion: 0.38
Nodes (5): Queries, Queries, Tx, New(), DBTX

### Community 485 - "WatchTask_Request"
Cohesion: 0.20
Nodes (3): file_tasks_proto_init(), init(), WatchTask_Request

### Community 487 - "CreateDeploy"
Cohesion: 0.11
Nodes (4): SizeCache, GetServiceGraph_Request, GetVervonomicon_Response, StopService

### Community 488 - "T17 — TopBar Widget"
Cohesion: 0.33
Nodes (5): Component, CSS, Props interface, T17 — TopBar Widget, What it looks like

### Community 489 - "T19 — KanbanBoard Widget"
Cohesion: 0.33
Nodes (5): Component, CSS, Props interface, T19 — KanbanBoard Widget, What it looks like

### Community 490 - "T20 — ServiceListView Widget"
Cohesion: 0.33
Nodes (5): Component, CSS, Props interface, T20 — ServiceListView Widget, What it looks like

### Community 491 - "T21 — NodeHealthList Widget"
Cohesion: 0.33
Nodes (5): Component, CSS, Props interface, T21 — NodeHealthList Widget, What it looks like

### Community 492 - "T24 — VCNPeerTable Widget"
Cohesion: 0.33
Nodes (5): Component, CSS, Props interface, T24 — VCNPeerTable Widget, What it looks like

### Community 493 - "T26 — ControlPlanePage (rebuild)"
Cohesion: 0.33
Nodes (5): Component, CSS, Data, T26 — ControlPlanePage (rebuild), What it looks like

### Community 494 - "T28 — VCNPage (rebuild)"
Cohesion: 0.33
Nodes (5): Component, CSS, Mock data, T28 — VCNPage (rebuild), What it looks like

### Community 495 - "DeleteEnvironment"
Cohesion: 0.33
Nodes (4): GetLoginServerUrl(), Context, Step, getLoginServerUrlStep

### Community 496 - "GetServiceResources"
Cohesion: 0.14
Nodes (4): file_service_api_proto_init(), init(), RemoveService_Response, RestartService_Response

### Community 497 - "fetch_by_api.go"
Cohesion: 0.40
Nodes (5): 2. PaaS platform, 3. UI redesign, 4. Genuinely open items carried forward from `docs/review/`, M1 — Core Platform, task-by-task, Velez Roadmap

### Community 498 - "VervServicesService"
Cohesion: 0.53
Nodes (4): Exec(), Context, Docker, execStep

### Community 499 - "renameContainerStep"
Cohesion: 0.38
Nodes (4): APIClient, Context, RenameContainer(), renameContainerStep

### Community 500 - "initKey"
Cohesion: 0.08
Nodes (30): AliveKeeper, AppInfo, Config, EnvironmentConfig, ServersConfig, SharedInstance, sharedInstanceCtxKey, NewSecurityManager() (+22 more)

### Community 501 - "CreateSmerdTaskPayload"
Cohesion: 0.13
Nodes (3): CreateSmerd_Request, CreateSmerdTaskPayload, UpgradeSmerdTaskPayload

### Community 502 - "factory-worker.md"
Cohesion: 0.50
Nodes (3): Inputs, Project rules (must follow), Workflow

### Community 503 - "Smerds management logic"
Cohesion: 0.50
Nodes (3): Engines, Smerds management logic, Updating / Restarting

### Community 504 - "Velez (lightweight node manager)"
Cohesion: 1.00
Nodes (3): Velez (lightweight node manager), master-actions RELEASE workflow, branch-push CI workflow

### Community 507 - "create_service.go"
Cohesion: 0.28
Nodes (5): Connection, DropSmerd_Request, DropSmerd_Response, Smerd, fakeContainerService

### Community 539 - "Option B: shell factory with direct Ollama API calls"
Cohesion: 0.33
Nodes (5): Component, CSS, Props interface, T08 — SectionLabel, What it looks like

### Community 540 - "Config subscription is commented out (handleConfigurationSubscription)"
Cohesion: 0.33
Nodes (3): DropSmerd_Request, DropSmerd_Response_Error, DropSmerdTaskPayload

### Community 541 - "GetService_Request"
Cohesion: 0.33
Nodes (4): CheckUpgradeIsAvailable(), Context, Step, checkUpgradeIsAvailableStep

### Community 548 - ".CreateNamespace"
Cohesion: 0.16
Nodes (11): Context, CreateVcnNamespace_Request, CreateVcnNamespace_Response, Impl, Context, ListVcnNamespaces_Request, ListVcnNamespaces_Response, Impl (+3 more)

### Community 550 - "createPgUserStep"
Cohesion: 0.12
Nodes (11): sortedFilePaths(), TestSortedFilePaths_Empty(), Docker, patchEnableStatefullDockerFields(), copyToVolumeHandler, dropSmerdHandler, Job, NamedJob (+3 more)

### Community 552 - "UpdateEnvironment"
Cohesion: 0.11
Nodes (4): UnknownFields, ListServices, RemoveService, StopService_Request

### Community 615 - "InitMaster"
Cohesion: 0.07
Nodes (5): Timestamp, CreateEnvironment_Response, Environment, ListEnvironments_Response, UpdateEnvironment_Response

### Community 619 - "ListPlugins"
Cohesion: 0.07
Nodes (10): MessageState, SizeCache, UnknownFields, ConnectSlave_Response, DeleteEnvironment_Response, InitMaster_Response, ListEnvironments_Request, ListPlugins (+2 more)

### Community 620 - ".DeleteEnvironment"
Cohesion: 0.40
Nodes (4): Context, Impl, DeleteEnvironment_Request, DeleteEnvironment_Response

### Community 679 - "fakeJobsEngine"
Cohesion: 0.67
Nodes (3): fakeJobsEngine, Context, VelezTask

### Community 694 - "AddMakoshRecord"
Cohesion: 0.06
Nodes (26): Configurator, ServiceDiscovery, VervClosedNetworkClient, clusterClients, MakoshBeAPIClient, MatreshkaBeAPIClient, RollMigration(), Client (+18 more)

### Community 733 - "StepsDialogHeader.tsx"
Cohesion: 0.05
Nodes (55): Plugin, VervPlugin, VervPluginState, VervPluginType, Dialog(), DialogManager, useDialog, queryClient (+47 more)

### Community 741 - "RemoveService"
Cohesion: 0.60
Nodes (4): APIClient, Context, ReadFromContainer(), WriteToContainer()

### Community 742 - "Jobs Migration — Open Questions"
Cohesion: 0.25
Nodes (7): ConnectServiceToVpn — resolved 2026-07-19, CopyToVolume (added during implementation), CopyToVolume — resolved 2026-07-19, Cross-cutting (all pipelines), EnableStatefullMode — resolved 2026-07-19, Jobs Migration — Open Questions, UpgradeSmerd — resolved 2026-07-19

### Community 746 - "drop_smerd.go"
Cohesion: 0.50
Nodes (3): Context, dropContainerJob, dropResultAccessor

### Community 747 - "ListPlugins_Request"
Cohesion: 0.13
Nodes (3): file_control_plane_api_proto_rawDescGZIP(), InitMaster_Request, ListNodes

### Community 748 - "Configurator"
Cohesion: 0.20
Nodes (6): ConfigurationPatch, Configurator, MatreshkaBeAPI_SubscribeOnChangesClient, MatreshkaBeAPIClient, New(), Configurator

### Community 753 - "Pipelines → Jobs Migration"
Cohesion: 0.40
Nodes (4): Migration checklist (repeat per pipeline), Pipelines → Jobs Migration, Status, Suggested pick-up order (easiest first)

### Community 757 - "VPN client key (Headscale auth key) storage in task context"
Cohesion: 0.50
Nodes (3): Decision, If this needs to change later, VPN client key (Headscale auth key) storage in task context

## Knowledge Gaps
- **1096 isolated node(s):** `status.sh script`, `go.vervstack.ru/Velez`, `UnsafeMatreshkaApiServer`, `UnsafeControlPlaneAPIServer`, `UnsafeServiceApiServer` (+1091 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **344 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `T` connect `internal/transport (grpcServer)` to `internal/transport (.ListEnvironments())`, `internal/clients (hardware_manager.go)`, `pkg/web (index.ts)`, `pkg/web (VervClosedNetworkPage.tsx)`, `internal/api (UnknownFields)`, `pkg/web (ControlPlanePage.tsx)`, `internal/storage (deployments)`, `pkg/web (PluginMatrix.tsx)`, `internal/service (list_test.go)`, `internal/transport (deploy_list.go)`, `createPgUserStep`, `internal/pipelines (ConnectServiceToVpn())`, `internal/clients (api_key_issue.go)`, `internal/pipelines (do_copy_to_volume.go)`, `internal/clients (NodeClients)`, `internal/pipelines (.UpgradeSmerd())`, `CreatePgUserForNode`, `internal/patterns (.EnableStatefullMode())`, `pkg/web (verv_closed_network.pb.ts)`, `tests/e2e (NewEnvironment())`, `internal/pipelines (fetch_from_container.go)`, `internal/api (UnknownFields)`, `internal/service (configurator.go)`, `.enrichServiceWithSmerdData`, `initKey`, `internal/cluster (service.go)`, `Exec`?**
  _High betweenness centrality (0.235) - this node is a cross-community bridge._
- **Why does `file_velez_common_proto_init()` connect `TasksStorage` to `GetServiceResources`, `internal/api (StopService_Request)`, `internal/api (control_plane_api.pb.go)`, `WatchTask_Request`?**
  _High betweenness centrality (0.153) - this node is a cross-community bridge._
- **Why does `New()` connect `ListPlugins_Request` to `internal/api (CreateDeploy_Request)`, `internal/pipelines (.UpgradeSmerd())`, `pkg/web (velez_api.pb.ts)`, `internal/api (GetServiceResources_Request)`?**
  _High betweenness centrality (0.128) - this node is a cross-community bridge._
- **Are the 71 inferred relationships involving `newFakeDocker()` (e.g. with `TestCreateScratchContainerJob_ContainerCreateError()` and `TestCreateScratchContainerJob_Rollback_NoContainerId_NoOp()`) actually correct?**
  _`newFakeDocker()` has 71 INFERRED edges - model-reasoned connections that need verification._
- **What connects `status.sh script`, `go.vervstack.ru/Velez`, `UnsafeMatreshkaApiServer` to the rest of the system?**
  _1111 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `internal/api (service_api.pb.gw.go)` be split into smaller, more focused modules?**
  _Cohesion score 0.058502799882110226 - nodes in this community are weakly interconnected._
- **Should `internal/api (velez_api.pb.gw.go)` be split into smaller, more focused modules?**
  _Cohesion score 0.06864035087719299 - nodes in this community are weakly interconnected._