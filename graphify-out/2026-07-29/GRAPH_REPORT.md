# Graph Report - Velez  (2026-07-29)

## Corpus Check
- 592 files · ~290,019 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 6762 nodes · 11515 edges · 734 communities (390 shown, 344 thin omitted)
- Extraction: 94% EXTRACTED · 6% INFERRED · 0% AMBIGUOUS · INFERRED: 647 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `6a199b83`
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
- Container_Healthcheck
- internal/api (ListPeers_Request)
- internal/api (StopService_Request)
- internal/api (Version_Response)
- GetServiceGraph
- tests/e2e (suite_control_plane_test.go)
- internal/clients (list_containers.go)
- internal/clients (Filter)
- Community 240
- selectiveFailDocker
- pkg/web (.GetService())
- pkg/web (MockDeployRow.tsx)
- internal/api (ConnectSlave_Response)
- .CreateNewDeploy
- internal/api (CreateVcnNamespace)
- internal/api (DeleteVcnNamespace_Response)
- internal/api (GetHardware_Request)
- internal/api (InitMaster_Response)
- internal/api (ListDeployments)
- internal/api (ListEnvironments)
- internal/api (ListPlugins_Request)
- internal/api (ListVcnNamespaces)
- internal/api (ListVcnNamespaces_Request)
- internal/api (MakeConnections_Response)
- Message
- ConnectSlave_Request
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
- EnablePlugin
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
- ListEnvironments_Response
- internal/transport (.GetService())
- internal/transport (.RestartService())
- internal/transport (.StopService())
- internal/transport (.GetVervonomicon())
- internal/transport (.ListPeers())
- internal/transport (.DeleteNamespace())
- RestartService
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
- stateManager
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
- pkg/web (ActivityPoint.tsx)
- Input.tsx
- pkg/web (Section.tsx)
- pkg/web (TextInput.tsx)
- ServicesStorage
- RegisterVelezAPIHandler
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
- prepareVervConfig
- file_control_plane_api_proto_rawDescGZIP
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
- EnablePlugin_Response
- fakeJobsStorage
- Pattern
- internal/clients (ns_list.go)
- .GetServiceEnvironments
- connectServiceToVpnHandler
- internal/storage (nodes.sql.go)
- go.mod (go.vervstack.ru/Velez)
- pkg/web (eslint.config.js)
- Community 422
- schemas/Networking.md (Verv Networking 101 (stub))
- Community 428
- Velez UI Redesign — Roadmap
- .ListServices
- T30 — AppsPage + AppCard (logical services view)
- Container
- .Do
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
- UpsertPluginParams
- WatchTask_Request
- .Do
- GetConfigNode_Response
- Project Factory — Agent Context
- Task 000 — Example Task Title
- Task 001 — Health check endpoint
- renameContainerStep
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
- .ContainerCreate
- dockerServiceDepsStorage
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
- GetService_Request
- db.go
- db.go
- prepareUpgradeVervConfigJob
- AddMakoshRecord
- UpsertPluginParams
- T17 — TopBar Widget
- T19 — KanbanBoard Widget
- T20 — ServiceListView Widget
- T21 — NodeHealthList Widget
- T24 — VCNPeerTable Widget
- T26 — ControlPlanePage (rebuild)
- T28 — VCNPage (rebuild)
- ValuesMapping.tsx
- InitMaster_Response
- fetch_by_api.go
- ContainerStats
- GetHardware
- GetVervonomicon
- CreateSmerdTaskPayload
- factory-worker.md
- Smerds management logic
- Velez (lightweight node manager)
- DataSourcesConfig
- React + TypeScript + Vite
- SingleFunc
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
- .Get
- Option B: shell factory with direct Ollama API calls
- Config subscription is commented out (handleConfigurationSubscription)
- GetLoginServerUrl
- internal/service/service_manager/container_manager/smerd_list.go
- internal/service/service_manager/container_manager/smerds_drop.go
- ControlPlanePage component
- NodeHealthList widget
- PluginMatrix widget
- ListDeployments
- ListVervPlugins_Request
- GetServiceResources
- ListServices
- DeployPage (existing deploy form)
- ListEnvironments_Response
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
- ServiceDependenciesStorage
- .WithTx
- GetServiceGraph
- StopService
- NewService
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
- InitMaster_Response
- StopService_Response
- pgStorage
- Configurator
- golangci-lint configuration
- selectiveFailDocker
- runner
- .List
- local.issues.md — known code issues
- smerds.ts mapping module (mapSmerdToServiceCard, mapSmerdToAppData, mapSmerdStatus, LOCAL_NODE)
- StopService
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
- NetworkTopologyMap widget
- VCNPeerRow (VCNPeerData type)
- VCNPeerTable widget
- VervClosedNetworkPage component
- Jobs Migration — Open Questions
- Pipelines → Jobs Migration
- EnableStatefullTaskPayload
- VPN client key (Headscale auth key) storage in task context

## God Nodes (most connected - your core abstractions)
1. `T` - 246 edges
2. `NodeClients` - 66 edges
3. `newFakeDocker()` - 61 edges
4. `file_service_api_proto_rawDescGZIP()` - 50 edges
5. `newFakeNodeClients()` - 41 edges
6. `useToaster` - 38 edges
7. `newFakeContainerAPI()` - 37 edges
8. `file_velez_api_proto_rawDescGZIP()` - 32 edges
9. `file_matreshka_api_proto_rawDescGZIP()` - 31 edges
10. `file_control_plane_api_proto_rawDescGZIP()` - 30 edges

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
- **Velez Configuration Environment Variable Schema** — config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_config_configyaml, config_configtemplate, docs_configuration [INFERRED 0.90]
- **Matreshka Config Mock Pattern** — tests_config_mocks_hello_world_matreshkaconfig, tests_config_mocks_velez_default_config_matreshkaconfig, schemas_smerds_matreshka [INFERRED 0.75]

## Communities (734 total, 344 thin omitted)

### Community 0 - "internal/api (service_api.pb.gw.go)"
Cohesion: 0.06
Nodes (87): ClientConnInterface, Context, CreateDeploy_Request, CreateDeploy_Response, CreateService_Request, CreateService_Response, GetService_Request, GetService_Response (+79 more)

### Community 1 - "internal/api (velez_api.pb.gw.go)"
Cohesion: 0.06
Nodes (75): AssembleConfig_Request, AssembleConfig_Response, BreakConnections_Request, BreakConnections_Response, ClientConnInterface, Context, CreateSmerd_Request, DropSmerd_Request (+67 more)

### Community 2 - "internal/api (matreshka_api.pb.go)"
Cohesion: 0.05
Nodes (17): file_matreshka_api_proto_init(), file_matreshka_api_proto_rawDescGZIP(), MessageState, SizeCache, UnknownFields, init(), CreateConfig, DeleteConfig (+9 more)

### Community 3 - "internal/api (verv_closed_network.pb.gw.go)"
Cohesion: 0.08
Nodes (55): ClientConnInterface, ConnectService_Request, ConnectService_Response, ConnectUser_Request, ConnectUser_Response, Context, CreateVcnNamespace_Request, CreateVcnNamespace_Response (+47 more)

### Community 4 - "internal/api (control_plane_api.pb.gw.go)"
Cohesion: 0.08
Nodes (50): ConnectSlave_Request, ConnectSlave_Response, _ControlPlaneAPI_ConnectSlave_Handler(), _ControlPlaneAPI_EnablePlugin_Handler(), _ControlPlaneAPI_ListEnvironments_Handler(), _ControlPlaneAPI_ListNodes_Handler(), _ControlPlaneAPI_ListPlugins_Handler(), ClientConnInterface (+42 more)

### Community 5 - "pkg/web (velez_api.pb.ts)"
Cohesion: 0.10
Nodes (27): connectToNetwork(), disconnectFromNetworks(), fromContainerNetwork(), ConfigTypePrefix, ContainerState, Context, EndpointSettings, InspectResponse (+19 more)

### Community 6 - "pkg/web (ServiceCard.tsx)"
Cohesion: 0.08
Nodes (37): AppCard(), AppCardProps, EnvChip(), EnvChipProps, FreezeChip(), IncidentChip(), MiniBar(), MiniBarProps (+29 more)

### Community 7 - "docs/paas (ServiceInfoPage component)"
Cohesion: 0.14
Nodes (13): 1. DeploymentsPage (`src/pages/deployments/DeploymentsPage.tsx`), 2. AppsPage (`src/pages/apps/AppsPage.tsx`), 3. SearchPage (`src/pages/search/SearchPage.tsx`), 4. ControlPlanePage (`src/pages/controlplane/ControlPlanePage.tsx`), Acceptance Criteria, Blocked dependency, Do NOT change, Files to create (+5 more)

### Community 8 - "internal/storage (Storage)"
Cohesion: 0.06
Nodes (35): ListNodesResponse, NodeBaseInfo, NodeStatus, IconButton(), IconButtonProps, isMetricAmber(), isMetricRed(), mapNodeStatus() (+27 more)

### Community 9 - "pkg/web (ServiceInfoPage.tsx)"
Cohesion: 0.04
Nodes (65): ServiceBaseInfo, CreateSmerdRequest, Smerd, SmerdStatus, Toast, Toaster, useToaster, Button() (+57 more)

### Community 11 - "internal/clients (create.go)"
Cohesion: 0.07
Nodes (29): DeploymentStatus, Badge(), BadgeProps, SectionLabel(), SectionLabelProps, StatCard(), StatCardProps, CodeBlock() (+21 more)

### Community 12 - "pkg/web (DeployWidget.tsx)"
Cohesion: 0.13
Nodes (18): classifyImage(), fromMatreshkaYamlToEvon(), fromYamlToEvon(), APIClient, ConfigFormat, Context, Docker, isPostgresByImageTags() (+10 more)

### Community 13 - "internal/clients (PortManager)"
Cohesion: 0.14
Nodes (18): Connect(), ConnectToContainer(), getAPIAddress(), Context, Docker, Client, Container, Docker (+10 more)

### Community 14 - "internal/storage (storage.go)"
Cohesion: 0.08
Nodes (10): Pointer, Querier, fakeClusterStorage, localStorage, stateManager, DeploymentsStorage, JobsStorage, NodesStorage (+2 more)

### Community 15 - "internal/pipelines (fetch_by_api.go)"
Cohesion: 0.12
Nodes (16): ConfigFormat, Connection, Container, ContainerHardware, ContainerHealthcheck, ContainerSettings, FileConfig, Image (+8 more)

### Community 16 - "internal/transport (Pipeliner)"
Cohesion: 0.20
Nodes (12): Context, Docker, Volume, writeFileToContainer(), containerIDAccessor, copyAPI, copyFileJob, copyToVolumeRequestAccessor (+4 more)

### Community 17 - "internal/transport (grpcServer)"
Cohesion: 0.06
Nodes (30): CMux, Cors, FS, HandlerFunc, Context, Listener, ServeMux, ServerOption (+22 more)

### Community 18 - "internal/api (EnumNumber)"
Cohesion: 0.08
Nodes (7): EnumDescriptor, EnumNumber, EnumType, ConfigFormat, NodeStatus, RestartPolicyType, Smerd_Status

### Community 19 - "pkg/web (index.ts)"
Cohesion: 0.07
Nodes (75): CopyToVolumeTaskPayload, T, TestCreateSmerdRequest_NilVervAndPlainStayNil(), TestCreateSmerdRequest_PlainSurvivesRoundTrip(), TestCreateSmerdRequest_SurvivesRoundTripEmbeddedInTaskPayload(), TestCreateSmerdRequest_VervAndPlainBothSurviveRoundTrip(), TestCreateSmerdRequest_VervSurvivesRoundTrip(), TestUpgradeSmerdRequest_SurvivesRoundTripEmbeddedInTaskPayload() (+67 more)

### Community 20 - "pkg/web (VervClosedNetworkPage.tsx)"
Cohesion: 0.19
Nodes (10): subscribeForConfigChangesStep, unsubscribeForConfigChangesStep, Context, SubscribeForConfigChanges(), Context, UnSubscribeForConfigChanges(), APIClient, Impl (+2 more)

### Community 22 - "internal/api (Context)"
Cohesion: 0.12
Nodes (21): ClientConnInterface, Context, ServiceRegistrar, UnaryServerInterceptor, Version_Request, Version_Response, _MatreshkaApi_CreateConfig_Handler(), _MatreshkaApi_DeleteConfig_Handler() (+13 more)

### Community 23 - "pkg/web (service_api.pb.ts)"
Cohesion: 0.03
Nodes (61): AboutService, Absent, BaseCreateDeployRequest, BaseGetServiceResponse, BoundResource, CreateDeploy, CreateDeployRequest, CreateDeployRequestUpgrade (+53 more)

### Community 24 - "internal/clients (cluster.go)"
Cohesion: 0.10
Nodes (13): TasksApi, GetHardwareResponse, ListSmerdsRequest, ListSmerdsResponse, SearchImagesResponse, VelezAPI, toProto(), DeploySmerd() (+5 more)

### Community 25 - "internal/storage (deployments)"
Cohesion: 0.12
Nodes (3): file_service_api_proto_rawDescGZIP(), CreateDeploy_Response, CreateService_Request

### Community 26 - "pkg/web (PluginMatrix.tsx)"
Cohesion: 0.31
Nodes (8): DeploymentInfo, DeploymentHistoryProps, DeployRow(), DeployRowProps, formatTimestamp(), getStatusClass(), parseImageTag(), TODO: replace with git commit hash when GetServiceGraph adds git field

### Community 27 - "internal/app (custom.go)"
Cohesion: 0.07
Nodes (23): Dialog(), DialogManager, useDialog, HeadscalePluginForm(), SimplePluginForm(), OpenWizardDialog(), TestContext, ScreenSkeletonLoader() (+15 more)

### Community 28 - "internal/clients (client.go)"
Cohesion: 0.09
Nodes (19): Docker, BoundResource, ContainerStats, ServiceMetrics, asciiSymbolsOnly(), APIClient, Context, CreateResponse (+11 more)

### Community 29 - "internal/api (MessageState)"
Cohesion: 0.08
Nodes (9): MessageState, SizeCache, UnknownFields, ConnectSlave, EnablePlugin_Response, InitMaster, ListEnvironments_Request, ListPlugins (+1 more)

### Community 30 - "CLAUDE.md (Velez CLAUDE.md project guide)"
Cohesion: 0.40
Nodes (6): makosh/headscale (VPN/network management), matreshka (external configuration service), Velez local state (VelezKey / local_state.json), config/config.yaml (production config), config/config_template.yaml, Matreshka Service Icon

### Community 31 - "internal/storage (node.go)"
Cohesion: 0.10
Nodes (18): ContainerManager, Configurator, Container, Context, Docker, Service, New(), Context (+10 more)

### Community 32 - "internal/service (services.go)"
Cohesion: 0.08
Nodes (23): 1. Ubuntu / Debian (apt), 2. Homebrew (macOS), 3. Chocolatey (Windows), Community repo vs. self-hosted feed, Config file placement, Docker's host-access requirements don't disappear on macOS, Does Velez support running as a native Windows service today?, Formula structure and the goreleaser nuance (+15 more)

### Community 33 - "internal/service (list_test.go)"
Cohesion: 0.18
Nodes (10): Connection, Context, DropSmerd_Request, DropSmerd_Response, ListSmerds_Request, ListSmerds_Response, Service, Smerd (+2 more)

### Community 34 - "internal/transport (deploy_list.go)"
Cohesion: 0.22
Nodes (7): Context, DeploymentStatus, ListDeployments_Request, ListDeployments_Response, Impl, VelezDeploymentStatus, toDeploymentStatus()

### Community 35 - "pkg/web (compilerOptions)"
Cohesion: 0.08
Nodes (24): compilerOptions, allowImportingTsExtensions, allowJs, allowSyntheticDefaultImports, esModuleInterop, forceConsistentCasingInFileNames, isolatedModules, jsx (+16 more)

### Community 37 - "internal/config (launch.go)"
Cohesion: 0.08
Nodes (30): AliveKeeper, AppInfo, Config, EnvironmentConfig, ServersConfig, SharedInstance, sharedInstanceCtxKey, NewSecurityManager() (+22 more)

### Community 38 - "internal/api (VervAppService)"
Cohesion: 0.09
Nodes (6): EnumDescriptor, EnumNumber, EnumType, DeploymentStatus, NodeType, VervAppService

### Community 39 - "pkg/web (services.ts)"
Cohesion: 0.23
Nodes (7): Context, Mutex, Queries, Tx, VelezTask, newTasksStorage(), tasks

### Community 40 - "internal/cluster (Context)"
Cohesion: 0.08
Nodes (14): ApiVersion_Request, ApiVersion_Response, disabledConfigurator, disabledServiceDiscovery, Context, ListEndpoints_Request, ListEndpoints_Response, UpsertEndpoints_Request (+6 more)

### Community 41 - "internal/pipelines (ConnectServiceToVpn())"
Cohesion: 0.10
Nodes (24): NewPortManager(), TestFetchSmerdConfigJob_SetEnv_CallerValueNotOverwritten(), TestFetchSmerdConfigJob_VervAndPlain_BothApply(), newFakeConfigurationService(), newFakeContainerService(), realPortManager(), strPtr(), TestCaptureOldContainerJob_InspectError() (+16 more)

### Community 42 - "internal/pipelines (ConfigurationService)"
Cohesion: 0.12
Nodes (11): FromContainerToRequest(), Context, NetworkBind, Smerd, CheckUpgradeIsAvailable(), Context, Step, upgradeSmerdHandler (+3 more)

### Community 43 - "internal/api (Smerd)"
Cohesion: 0.11
Nodes (3): Timestamp, Smerd, Volume

### Community 44 - "internal/api (CreateSmerd_Request)"
Cohesion: 0.10
Nodes (7): Container_Hardware, Container_Healthcheck, Container_Settings, FileConfig, MatreshkaConfigSpec, RestartPolicy, CreateSmerd_Request

### Community 45 - "internal/pipelines (LaunchSmerd)"
Cohesion: 0.16
Nodes (13): Context, CreateRequest, Docker, clientKeyAccessor, connectServiceToVpnRequestAccessor, createSidecarContainerJob, getClientKeyJob, getLoginServerURLJob (+5 more)

### Community 46 - "internal/api (VervPluginType)"
Cohesion: 0.31
Nodes (7): Context, ListPlugins_Response, Storage, listInactivePlugins(), NewPluginService(), Plugin, pluginService

### Community 47 - "pkg/web (Button.tsx)"
Cohesion: 0.10
Nodes (24): NullFloat64, NullInt32, NullInt64, NullRawMessage, NullString, NullTime, Time, Value (+16 more)

### Community 48 - "pkg/web (creds.ts)"
Cohesion: 0.06
Nodes (41): EnvCard(), EnvCardProps, ServiceEnvironment, ServiceGraphNode, ServiceResource, VervonomiconDocs, IsStatefullModeEnabled(), ListEnvironmentsQuery() (+33 more)

### Community 49 - "internal/pipelines (do_copy_to_volume.go)"
Cohesion: 0.05
Nodes (39): b64, b64Decode(), b64Encode(), fetchStreamingRequest(), FlattenedRequestPayload, flattenRequestPayload(), getNewLineDelimitedJSONDecodingStream(), getNotifyEntityArrivalSink() (+31 more)

### Community 50 - "internal/clients (NodeClients)"
Cohesion: 0.07
Nodes (26): AssembleConfig, AssembleConfigRequest, AssembleConfigResponse, BreakConnections, BreakConnectionsRequest, BreakConnectionsResponse, CreateSmerd, DropSmerd (+18 more)

### Community 51 - "internal/pipelines (.UpgradeSmerd())"
Cohesion: 0.09
Nodes (26): CreateServiceRequest, ConnectionHealth, ConnectionHealthStatus, useConnectionHealth(), MainLayout(), NAV_TO_ROUTE, NavId, ROUTE_TO_NAV (+18 more)

### Community 52 - "internal/api (SubscribeOnChanges_Response)"
Cohesion: 0.09
Nodes (7): BidiStreamingServer, BidiStreamingClient, ServerStream, _MatreshkaApi_SubscribeOnChanges_Handler(), BidiStreamingClient, SubscribeOnChanges_Request, SubscribeOnChanges_Response

### Community 53 - "tests/e2e (T)"
Cohesion: 0.14
Nodes (3): file_matreshka_common_proto_rawDescGZIP(), Sort, Sort_Type

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
Cohesion: 0.18
Nodes (10): countTotal(), fromStorageToDomainService(), Context, Querier, SelectBuilder, Service, ServiceBaseInfo, VelezService (+2 more)

### Community 59 - "internal/api (control_plane_api.pb.go)"
Cohesion: 0.12
Nodes (9): EnableHeadscaleServer_DeployHeadscaleConfig, EnableHeadscaleServer_ExternalHeadscaleConnection, file_control_plane_api_proto_init(), init(), EnableHeadscaleServer, EnableHeadscaleServer_DeployConfig, EnableHeadscaleServer_ExternalConnect, EnablePlugin_Request_HeadscaleServer (+1 more)

### Community 60 - "pkg/web (dependencies)"
Cohesion: 0.10
Nodes (24): NullFloat64, NullInt32, NullInt64, NullRawMessage, NullString, NullTime, Time, Value (+16 more)

### Community 61 - "internal/api (ServiceEnvironmentInfo)"
Cohesion: 0.10
Nodes (3): Timestamp, GetServiceEnvironments_Response, ServiceEnvironmentInfo

### Community 62 - "internal/clients (state.go)"
Cohesion: 0.12
Nodes (11): firstNotEmptyKey(), Once, readStateFromPath(), writeKey(), fakeStateManager, ClusterState, Headscale, Manager (+3 more)

### Community 63 - "pkg/web (fetch.pb.ts)"
Cohesion: 0.13
Nodes (14): CreateDeploymentParams, CreateSpecificationParams, GetSpecificationByIdRow, UpdateDeploymentStatusParams, Context, Queries, Tx, newDeploymentsStorage() (+6 more)

### Community 64 - "pkg/web (ServiceService)"
Cohesion: 0.10
Nodes (18): APIClient, Context, Docker, FileConfig, MatreshkaConfigSpec, copyToContainerJob, createContainerJob, createSmerdHandler (+10 more)

### Community 65 - "pkg/web (HomePage.tsx)"
Cohesion: 0.12
Nodes (5): MessageState, UnknownFields, isPatch_Patch, Paging, Patch

### Community 66 - "internal/storage (docker_impl.go)"
Cohesion: 0.17
Nodes (12): Credentials, ls, Settings, SettingsBase, storeToLocalStorage(), useSettings(), GetInitReq(), InitReq (+4 more)

### Community 68 - "internal/api (VervPlugin_State)"
Cohesion: 0.09
Nodes (6): EnumDescriptor, EnumNumber, EnumType, VervPlugin, VervPlugin_State, VervPluginType

### Community 69 - "pkg/web (service.ts)"
Cohesion: 0.13
Nodes (15): Deployment, DeploymentList, DeploymentSpecification, ListDeploymentsReq, Paging, Time, VelezDeploymentStatus, Context (+7 more)

### Community 70 - "pkg/web (verv_closed_network.pb.ts)"
Cohesion: 0.10
Nodes (38): proto(), TestStartSidecarContainerJob_ContainerStartError(), TestStartSidecarContainerJob_NoContainerId_Error(), TestStartSidecarContainerJob_Rollback_StopsContainer(), TestStartSidecarContainerJob_Success(), TestStartLoaderContainerJob_ContainerStartError(), TestStartLoaderContainerJob_NoContainerId_Error(), TestStartLoaderContainerJob_Rollback_NoContainerId_NoOp() (+30 more)

### Community 71 - "internal/storage (models.go)"
Cohesion: 0.10
Nodes (24): NullVelezDeploymentStatus, NullVelezJobStatus, NullVelezTaskStatus, VelezDeployment, VelezDeploymentSpecification, VelezDeploymentStatus, VelezJob, VelezJobStatus (+16 more)

### Community 72 - "internal/domain (service.go)"
Cohesion: 0.07
Nodes (26): Context, Duration, VelezTask, NewEngine(), TestEngine_EnqueueDedupesConcurrentCalls(), TestEngine_EnqueueReturnsExistingFailedTaskInstead(), Context, GetServiceEnvironments_Request (+18 more)

### Community 73 - "tests/e2e (NewEnvironment())"
Cohesion: 0.05
Nodes (46): AssembleConfigJobSuite, AssembleConfigSuite, LifecycleSuite, StateOpt, TestEnvironment, TestEnvOpt, UpgradeSmerdSuite, VpnSuite (+38 more)

### Community 74 - "internal/api (velez_common.pb.go)"
Cohesion: 0.05
Nodes (30): VervClosedNetworkClient, clusterClients, Client, Configurator, Context, ServiceDiscovery, Setup(), Context (+22 more)

### Community 75 - "internal/storage (models.go)"
Cohesion: 0.10
Nodes (24): NullFloat64, NullInt32, NullInt64, NullRawMessage, NullString, NullTime, Time, Value (+16 more)

### Community 76 - "internal/storage (models.go)"
Cohesion: 0.10
Nodes (24): NullFloat64, NullInt32, NullInt64, NullRawMessage, NullString, NullTime, Time, Value (+16 more)

### Community 77 - "internal/storage (models.go)"
Cohesion: 0.10
Nodes (24): NullFloat64, NullInt32, NullInt64, NullRawMessage, NullString, NullTime, Time, Value (+16 more)

### Community 78 - "internal/storage (models.go)"
Cohesion: 0.10
Nodes (24): NullFloat64, NullInt32, NullInt64, NullRawMessage, NullString, NullTime, Time, Value (+16 more)

### Community 79 - "internal/storage (models.go)"
Cohesion: 0.10
Nodes (24): NullFloat64, NullInt32, NullInt64, NullRawMessage, NullString, NullTime, Time, Value (+16 more)

### Community 81 - "internal/pipelines (fetch_from_container.go)"
Cohesion: 0.13
Nodes (9): AppConfig, ConfigMeta, ConfigurationPatch, ConfigFormat, ConfigTypePrefix, Configurator, Context, Configurator (+1 more)

### Community 82 - "internal/storage (service_dependencies.go)"
Cohesion: 0.08
Nodes (24): Custom, ClusterClients, Impl, App, Context, Listener, New(), App (+16 more)

### Community 86 - "internal/api (Format)"
Cohesion: 0.25
Nodes (8): AutoUpgrade, APIClient, Context, Duration, Once, Summary, imageNameWithoutTag(), New()

### Community 87 - "internal/api (service_api.pb.go)"
Cohesion: 0.09
Nodes (14): Closable, VcnNamespace, createNamespaceRequest, createNamespaceResponse, getNamespaceResponse, listNamespacesResponse, Context, Client (+6 more)

### Community 88 - "internal/api (Message)"
Cohesion: 0.14
Nodes (17): RollMigration(), getExposedPgPort(), Context, InspectResponse, containerInspectAPI, createPgContainerJob, createPgUserJob, createSchemaAndMigrateJob (+9 more)

### Community 90 - "internal/api (ListDeployments_Request)"
Cohesion: 0.13
Nodes (3): Paging, ListDeployments_Request, ListServices_Request

### Community 91 - "internal/api (SizeCache)"
Cohesion: 0.18
Nodes (11): AboutService, CreateServiceReq, RemoveServiceReq, Service, ServiceBaseInfo, ServiceList, DeploymentStatus, ServiceBaseInfo (+3 more)

### Community 92 - "internal/api (UnknownFields)"
Cohesion: 0.24
Nodes (5): HelloWorldClusterSuite, APIClient, Context, Smerd, Suite

### Community 94 - "pkg/web (Input.tsx)"
Cohesion: 0.11
Nodes (17): LaunchSmerd, LaunchSmerdResult, CreateSmerd_Request, pipeliner, FetchConfig(), APIClient, Context, Healthcheck() (+9 more)

### Community 98 - "internal/service (configurator.go)"
Cohesion: 0.21
Nodes (6): Tx, Docker, Storage, VervService, New(), TxManager

### Community 99 - "internal/api (Message)"
Cohesion: 0.25
Nodes (3): file_velez_api_proto_init(), init(), CreateSmerd

### Community 101 - "internal/api (PatchConfig_Request)"
Cohesion: 0.20
Nodes (7): Connection, DropSmerd_Request, DropSmerd_Response, ListSmerds_Request, ListSmerds_Response, Smerd, fakeContainerService

### Community 103 - "internal/api (SaveConfig_Request)"
Cohesion: 0.07
Nodes (22): CopyToContainerOptions, CreateOptions, APIClient, Context, CreateResponse, EndpointSettings, ExecOptions, HostConfig (+14 more)

### Community 104 - "pkg/web (control_plane.ts)"
Cohesion: 0.10
Nodes (18): CopyToVolumeRequest, pipeliner, NewCopyToVolumeRunner(), pipeliner, Runner, DropContainerStep(), Context, Docker (+10 more)

### Community 107 - "tests/e2e (helper.go)"
Cohesion: 0.23
Nodes (7): Context, Queries, newServiceDependenciesStorage(), BoundResource, Context, wrapPgErr(), serviceDependenciesStorage

### Community 109 - "internal/api (Sort_Type)"
Cohesion: 0.10
Nodes (5): EnumDescriptor, EnumNumber, EnumType, ConfigType, Format

### Community 110 - "internal/api (BreakConnections_Request)"
Cohesion: 0.15
Nodes (3): Connection, BreakConnections_Request, MakeConnections_Request

### Community 111 - "internal/cluster (local.go)"
Cohesion: 0.16
Nodes (11): Context, Docker, ListEndpoints_Request, ListEndpoints_Response, ServiceDiscovery, UpsertEndpoints_Request, UpsertEndpoints_Response, Version_Request (+3 more)

### Community 114 - "Community 114"
Cohesion: 0.18
Nodes (10): CSS, EnvChip, EnvChip, FreezeChip, FreezeChip, IncidentChip, IncidentChip, Notes (+2 more)

### Community 115 - "internal/api (velez_api.pb.go)"
Cohesion: 0.15
Nodes (6): file_service_api_proto_init(), CreateDeploy_Request_Upgrade, init(), CreateDeploy_Request, GetService_Response_VervService, isCreateDeploy_Request_Specification

### Community 116 - "internal/api (Message)"
Cohesion: 0.04
Nodes (61): Absent, BaseEnableHeadscaleServer, BaseEnablePluginRequest, ConnectSlave, ConnectSlaveRequest, ConnectSlaveResponse, ControlPlaneAPI, EnableHeadscaleServer (+53 more)

### Community 117 - "internal/clients (Filter)"
Cohesion: 0.19
Nodes (5): Args, New(), Filter, health, status

### Community 118 - "internal/cluster (service.go)"
Cohesion: 0.22
Nodes (7): ListEndpoints_Request, ListEndpoints_Response, UpsertEndpoints_Request, UpsertEndpoints_Response, Version_Request, Version_Response, fakeServiceDiscovery

### Community 119 - "internal/api (CreateConfig_Request)"
Cohesion: 0.19
Nodes (13): Context, CreateSmerd_Request, Handler, ServerStreamingServer, TaskStatus, VelezTask, VelezTaskStatus, WatchTask_Request (+5 more)

### Community 124 - ".claude/factory (factory.sh)"
Cohesion: 0.35
Nodes (11): check_deps(), check_ollama(), err(), list_tasks(), log(), main(), ok(), pick_task() (+3 more)

### Community 125 - "Community 125"
Cohesion: 0.16
Nodes (18): App (Velez Dashboard reference mockup), Badge component (mockup), ControlPlaneTab component (mockup), DeployDialog (multi-step deploy modal), DeploymentsTab (kanban + list view), EnvChip component (mockup), FreezChip (release-freeze chip, mockup), IncidentChip component (mockup) (+10 more)

### Community 126 - "tests/e2e (helper_environment.go)"
Cohesion: 0.25
Nodes (8): Context, Mutex, Queries, Tx, VelezJob, jobKey(), newJobsStorage(), jobs

### Community 127 - "internal/api (ConfigBase)"
Cohesion: 0.08
Nodes (4): Message, CreateConfig_Request, CreateConfig_Response, Version_Response

### Community 129 - "tests/e2e (TestEnvironment)"
Cohesion: 0.26
Nodes (6): Context, Duration, Once, Ticker, VelezTask, taskWorker

### Community 131 - "internal/transport (.CreateDeploy())"
Cohesion: 0.29
Nodes (7): CreateSmerd_Request, Context, CreateDeploy_Request, CreateDeploy_Request_Upgrade_, CreateDeploy_Response, Impl, CreateDeploy_Request_New

### Community 132 - "internal/api (MessageState)"
Cohesion: 0.09
Nodes (7): MessageState, SizeCache, UnknownFields, Connection, Container, Image, PlainConfigSpec

### Community 133 - "internal/pipelines (Step)"
Cohesion: 0.27
Nodes (7): createContainerStep, Create(), APIClient, Context, CreateRequest, Docker, Step

### Community 135 - "internal/storage (services.go)"
Cohesion: 0.23
Nodes (11): containerStateToDeploymentStatus(), containerStateToString(), countRunningServices(), Context, DeploymentStatus, Docker, Service, ServiceBaseInfo (+3 more)

### Community 136 - "pkg/web (ResourcesSection.tsx)"
Cohesion: 0.09
Nodes (10): Pointer, NewContainer(), Container, Docker, fakeNodeClients, Docker, HardwareManager, StateManager (+2 more)

### Community 140 - "Community 140"
Cohesion: 0.16
Nodes (11): Context, CreateVcnNamespace_Request, CreateVcnNamespace_Response, Impl, Context, ListVcnNamespaces_Request, ListVcnNamespaces_Response, Impl (+3 more)

### Community 142 - "internal/pipelines (create.go)"
Cohesion: 0.16
Nodes (4): SizeCache, Timestamp, ConfigBase, ConfigInfo

### Community 144 - "internal/api (file_velez_api_proto_rawDescGZIP())"
Cohesion: 0.11
Nodes (13): envContainer, pgStorage, staticStorage, Context, Pointer, NewContainer(), Context, NewPg() (+5 more)

### Community 147 - "internal/api (SizeCache)"
Cohesion: 0.15
Nodes (13): dependencies, classnames, framer-motion, @microlink/react-json-view, react, react-dom, react-router-dom, react-tooltip (+5 more)

### Community 151 - "internal/transport (network.go)"
Cohesion: 0.22
Nodes (8): BreakConnections_Request, BreakConnections_Response, Connection, Context, MakeConnections_Request, MakeConnections_Response, Impl, toConnection()

### Community 152 - "pkg/web (ControlPlanePage.tsx)"
Cohesion: 0.07
Nodes (50): EnableStatefullTaskPayload, App, Checkpoint(), Context, TestCheckpoint_FailFastWhenFailed(), TestCheckpoint_RunsAndPersistsContextOnSuccess(), TestCheckpoint_SkipsWhenDone(), NewCreateServiceHandler() (+42 more)

### Community 153 - "internal/api (EnableStatefullCluster)"
Cohesion: 0.11
Nodes (4): EnablePlugin_Request, EnablePlugin_Request_StatefullCluster, EnableStatefullCluster, isEnablePlugin_Request_Payload

### Community 156 - "internal/transport (VcnNamespace)"
Cohesion: 0.10
Nodes (18): New(), NewTxManager(), Queries, newDeploymentsStorage(), Queries, newJobsStorage(), Queries, newServiceResourcesStorage() (+10 more)

### Community 157 - "internal/api (ListConfigs_Request)"
Cohesion: 0.12
Nodes (3): Paging, ListConfigs_Request, ListConfigs_Response

### Community 158 - "internal/api (file_control_plane_api_proto_rawDescGZIP())"
Cohesion: 0.14
Nodes (25): ConnectServiceToVpnTaskPayload, NewConnectServiceToVpnHandler(), connectServiceToVpnTask(), VelezTask, patchStartAPIField(), TestAddMakoshRecordJob_Success(), TestAddMakoshRecordJob_UpsertError(), TestConnectServiceToVpnHandler_Action() (+17 more)

### Community 159 - "tests/e2e (suite_hello_world_cluster_test.go)"
Cohesion: 0.08
Nodes (14): ControlPlaneSuite, EnableStatefullSuite, Context, runner[T], Test_AssembleConfigJob(), Test_AssembleConfig(), Suite, Test_ControlPlane() (+6 more)

### Community 161 - "pkg/web (VcnApi)"
Cohesion: 0.08
Nodes (26): Code, fetchConfigStep, getConfigFromContainerStep, copyToContainerStep, ConfigMount, FileMountPoint, RespError, APIClient (+18 more)

### Community 162 - "pkg/web (compilerOptions)"
Cohesion: 0.22
Nodes (8): compilerOptions, allowSyntheticDefaultImports, composite, module, moduleResolution, skipLibCheck, strict, include

### Community 165 - "internal/cluster (Synchronizer)"
Cohesion: 0.25
Nodes (5): Synchronizer, Client, Context, MatreshkaBeAPI_SubscribeOnChangesClient, New()

### Community 166 - "internal/pipelines (create_and_inspect.go)"
Cohesion: 0.27
Nodes (11): NodeBaseInfo, NodeStatusFromLastOnline(), Test_NodeStatusFromLastOnline_DisabledNodeIsOffline(), Test_NodeStatusFromLastOnline_FreshHeartbeatIsOnline(), Test_NodeStatusFromLastOnline_PastOfflineThresholdIsOffline(), Test_NodeStatusFromLastOnline_StaleUnderOfflineThresholdIsDegraded(), Test_ToBasicNodeInfo_PassesThroughRegionUnchanged(), Test_ToBasicNodeInfo_PassesThroughUsageFieldsUnchanged() (+3 more)

### Community 167 - "internal/domain (images.go)"
Cohesion: 0.31
Nodes (7): ImageListRequest, ImageSearchRequest, APIClient, Context, Image, ListImages(), SearchImages()

### Community 168 - "internal/cluster (DisabledVcnImpl)"
Cohesion: 0.10
Nodes (14): ConnectServiceToVcn, IssueClientKey, ListVpnNamespaces, RegisterVcnNodeReq, SetupHeadscaleRequest, SetupHeadscaleResponse, Context, Client (+6 more)

### Community 170 - "internal/clients (api_key_issue.go)"
Cohesion: 0.23
Nodes (8): Context, NodeBaseInfo, Querier, SelectBuilder, newNodeStorage(), scanNode(), nodeStorage, Scannable

### Community 171 - "internal/api (matreshka_common.pb.go)"
Cohesion: 0.18
Nodes (4): fromApiNodes(), Configurator, Context, Node

### Community 173 - "internal/api (verv_closed_network.pb.go)"
Cohesion: 0.25
Nodes (3): file_verv_closed_network_proto_init(), init(), ConnectUser

### Community 174 - "internal/storage (service_dependencies.sql.go)"
Cohesion: 0.33
Nodes (5): Context, Queries, GetServiceCallersRow, GetServiceDependenciesRow, UpsertServiceDependencyParams

### Community 175 - "pkg/web (package.json)"
Cohesion: 0.13
Nodes (6): MessageState, SizeCache, UnknownFields, CreateServiceTaskPayload, WatchTask, WatchTask_Request

### Community 176 - "internal/api (Connection)"
Cohesion: 0.12
Nodes (4): SizeCache, CreateDeploy, GetServiceEnvironments, GetServiceMetrics

### Community 177 - "internal/api (Container_Hardware)"
Cohesion: 0.53
Nodes (4): keyIssuer, Context, Docker, issueNewAPIKey()

### Community 178 - "internal/api (Container_Healthcheck)"
Cohesion: 0.18
Nodes (9): PluginBaseInfo, DeploymentStatus, VervPlugin_State, calculatePluginState(), Context, Querier, VervPlugin_State, newPluginsStorage() (+1 more)

### Community 180 - "internal/api (Image)"
Cohesion: 0.12
Nodes (4): MessageState, CreateService, RemoveService, RestartService

### Community 183 - "internal/api (SearchImageItem)"
Cohesion: 0.18
Nodes (3): UnknownFields, AssembleConfig, DropSmerd

### Community 188 - "internal/domain (vcn.go)"
Cohesion: 0.22
Nodes (8): Backlog (suggested order — confirm before starting if you'd rather reorder), Context, Ground rules (apply to every task below, every agent), Open decisions (need explicit user sign-off before the affected work starts), Per-pipeline recipe (repeat for each row in the backlog), Pipeline #1 special case: `CreateSmerd` streaming pilot, Plan: Cut Pipelines Over to the Postgres/Storage-Based Jobs Engine, Progress log

### Community 193 - "internal/patterns (.EnableStatefullMode())"
Cohesion: 0.42
Nodes (3): CreateSmerd_Request, imageExposedPortsAccessor, prepareSmerdVervConfigJob

### Community 194 - "fetchSmerdConfigJob"
Cohesion: 0.14
Nodes (15): Checkbox(), CheckboxProps, InfoMark(), InfoMarkProps, PortMapping(), PortMappingProps, getLinkToPort(), CreateSmerdReq (+7 more)

### Community 195 - "internal/api (GetConfigNode_Request)"
Cohesion: 0.20
Nodes (9): Artel "Tract", Common ground (both systems converged independently — kept in the merge), Divergent decisions — resolved, Open trade-off to confirm before implementation, Per-system deep dive, Sources, Synthesized target architecture, Task/Job Engine — Velez vs. Artel Comparison & Merge Direction (+1 more)

### Community 196 - "internal/api (GetConfigNode_Response)"
Cohesion: 0.09
Nodes (20): Agent code of conduct, API (Proto definitions in `api/grpc/`), Architecture, Building, Code Generation, Code Style, Commands, Configuration (+12 more)

### Community 197 - "UpsertPluginParams"
Cohesion: 0.11
Nodes (4): UnknownFields, GetServiceResources, GetVervonomicon_Request, GetVervonomicon_Response

### Community 208 - ".claude/factory (review.sh)"
Cohesion: 0.52
Nodes (5): err(), log(), ok(), review_task(), review.sh script

### Community 209 - "internal/storage (db.go)"
Cohesion: 0.38
Nodes (5): DBTX, Queries, Queries, Tx, New()

### Community 213 - "internal/clients (.doApiRequest())"
Cohesion: 0.43
Nodes (4): Context, Client, Request, Response

### Community 215 - "Context"
Cohesion: 0.15
Nodes (12): Context, NullRawMessage, NullString, NullTime, Queries, VelezTask, VelezTaskStatus, ClaimTaskParams (+4 more)

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

### Community 224 - "internal/api (AssembleConfig_Response)"
Cohesion: 0.20
Nodes (3): file_velez_common_proto_init(), init(), Paging

### Community 225 - "internal/api (CreateService_Request)"
Cohesion: 0.33
Nodes (5): Component, CSS, Props interface, T15 — CodeBlock, What it looks like

### Community 228 - "internal/api (GetService_Request)"
Cohesion: 0.16
Nodes (14): ConnectToNetworkRequest, ConnectToNetwork(), CreateNetwork(), DisconnectFromNetworks(), APIClient, Context, EndpointSettings, APIClient (+6 more)

### Community 229 - "internal/api (GetServiceEnvironments_Request)"
Cohesion: 0.13
Nodes (10): ClusterStateManager, ClusterStateManagerContainer, Configurator, ServiceDiscovery, MakoshBeAPIClient, MatreshkaBeAPIClient, Storage, NewContainer() (+2 more)

### Community 230 - "internal/api (GetServiceMetrics_Request)"
Cohesion: 0.24
Nodes (7): Context, Docker, newPluginsStorage(), Docker, Storage, New(), dockerPluginsStorage

### Community 231 - "internal/api (GetServiceResources_Request)"
Cohesion: 0.05
Nodes (41): HealthConfig, FromCommand(), FromHealthcheck(), Container_Healthcheck, FromDockerEnv(), ToDockerEnv(), FromVolume(), Container_Settings (+33 more)

### Community 236 - "GetServiceGraph"
Cohesion: 0.19
Nodes (6): GetVcnAuthKeyReq, VcnAuthKey, Context, Client, Context, DisabledVcnImpl

### Community 237 - "tests/e2e (suite_control_plane_test.go)"
Cohesion: 0.13
Nodes (5): file_tasks_proto_rawDescGZIP(), EnumDescriptor, EnumNumber, EnumType, TaskStatus_Status

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
Cohesion: 0.16
Nodes (11): ListNodesReq, NodeBaseInfo, NodesList, NodeBaseInfo, Paging, Time, Context, Context (+3 more)

### Community 242 - "pkg/web (.GetService())"
Cohesion: 0.40
Nodes (4): Context, RemoveService_Request, RemoveService_Response, Impl

### Community 243 - "pkg/web (MockDeployRow.tsx)"
Cohesion: 0.22
Nodes (6): Configurator, Context, Configurator, Context, AppendPrefix(), ConfigTypePrefix

### Community 244 - "internal/api (ConnectSlave_Response)"
Cohesion: 0.16
Nodes (10): VelezJob, Context, Queries, NullString, VelezJob, fakeJobsStorage, CreateRunningJobParams, FinishJobParams (+2 more)

### Community 245 - ".CreateNewDeploy"
Cohesion: 0.38
Nodes (4): CreateDeployReq, UpgradeDeployReq, Context, VervService

### Community 248 - "internal/api (GetHardware_Request)"
Cohesion: 0.10
Nodes (15): sortedFilePaths(), TestSortedFilePaths_Empty(), Context, Docker, patchEnableStatefullDockerFields(), Docker, copyToVolumeHandler, dropContainerJob (+7 more)

### Community 250 - "internal/api (ListDeployments)"
Cohesion: 0.18
Nodes (13): Port, Volume, ValuesPair(), VolumeMappingProps, Bind, fromProto(), fromProtoPort(), fromProtoPorts() (+5 more)

### Community 251 - "internal/api (ListEnvironments)"
Cohesion: 0.22
Nodes (3): DropSmerd_Request, DropSmerd_Response_Error, DropSmerdTaskPayload

### Community 253 - "internal/api (ListPlugins_Request)"
Cohesion: 0.39
Nodes (6): ServiceBaseInfo, TestToServiceBaseInfoList(), TestToServiceBaseInfoWithEmptyFields(), TestToServiceBaseInfoWithEnrichedFields(), toServiceBaseInfo(), toServiceBaseInfoList()

### Community 257 - "internal/api (MakeConnections_Response)"
Cohesion: 0.33
Nodes (4): DeploymentFiltersProps, ENV_OPTIONS, STATUS_OPTIONS, ViewMode

### Community 258 - "Message"
Cohesion: 0.28
Nodes (6): Context, ServiceBaseInfo, VervService, mapSmerdStatus(), TestMapSmerdStatus(), Smerd_Status

### Community 260 - "ListNodes"
Cohesion: 0.33
Nodes (5): entry, ignore, project, $schema, knip

### Community 262 - "internal/api (UpgradeSmerd_Response)"
Cohesion: 0.18
Nodes (3): SizeCache, BreakConnections, BreakConnections_Response

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
Cohesion: 0.40
Nodes (4): Context, Impl, ListEnvironments_Request, ListEnvironments_Response

### Community 269 - "internal/transport (.ListPlugins())"
Cohesion: 0.40
Nodes (4): Context, Impl, ListPlugins_Request, ListPlugins_Response

### Community 271 - "internal/clients (hardware_manager.go)"
Cohesion: 0.21
Nodes (9): Manager, Context, GetHardware_Response, Mutex, Time, New(), TestManager_GetHardware_CachesWithinTTL(), TestManager_GetHardware_IsRunningInContainer() (+1 more)

### Community 275 - "internal/clients (pull_image.go)"
Cohesion: 0.40
Nodes (4): APIClient, Context, Image, PullImage()

### Community 277 - "VelezAPI"
Cohesion: 0.19
Nodes (9): ServerStream, ServiceRegistrar, RegisterTasksApiServer(), _TasksApi_CreateSmerdStream_Handler(), _TasksApi_WatchTask_Handler(), ServiceRegistrar, TasksApiServer, UnimplementedTasksApiServer (+1 more)

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

### Community 289 - "RestartService"
Cohesion: 0.15
Nodes (12): AssembleConfigTaskPayload, ConnectServiceToVpnTaskPayload, CopyToVolumeTaskPayload, CreateServiceTaskPayload, CreateSmerdTaskPayload, DropSmerdTaskPayload, EnableStatefullTaskPayload, TaskStatus (+4 more)

### Community 290 - "portManagerImpl"
Cohesion: 0.25
Nodes (5): file_matreshka_common_proto_init(), init(), Patch_Delete, Patch_Rename, Patch_UpdateValue

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

### Community 308 - "stateManager"
Cohesion: 0.27
Nodes (7): InitReq, Errors, GrpcError, parseGrpcError(), ServiceError, ApiService, withRetries()

### Community 310 - "PostgreSQL Database"
Cohesion: 0.67
Nodes (3): PostgreSQL Database, Database Pixel Icon, PostgreSQL Service Icon

### Community 311 - "CreatePgUserForNode"
Cohesion: 0.27
Nodes (7): Context, VelezTask, Test_CreateSmerdStream(), Test_CreateSmerdStream_EnqueueError(), enqueueCall, fakeJobsEngine, watchCall

### Community 319 - "Community 319"
Cohesion: 0.67
Nodes (3): Arrow Forward Icon, Open In Tab Icon, Rocket Icon

### Community 321 - "Input.tsx"
Cohesion: 0.15
Nodes (7): InputProps, StyleProps, KeyValueProps, PlainMapProps, Search(), SearchParam, SearchProps

### Community 322 - "pkg/web (Section.tsx)"
Cohesion: 0.43
Nodes (5): Context, Docker, InspectResponse, PrepareImage(), stepPrepareImage

### Community 323 - "pkg/web (TextInput.tsx)"
Cohesion: 0.40
Nodes (4): name, private, trustedDependencies, type

### Community 324 - "ServicesStorage"
Cohesion: 0.20
Nodes (5): Context, createServiceHandler, serviceNameAccessor, upsertServiceJob, ServicesStorage

### Community 325 - "RegisterVelezAPIHandler"
Cohesion: 0.25
Nodes (6): FromPaging(), Paging, Context, Impl, ListNodes_Request, ListNodes_Response

### Community 340 - "pkg/web (eslint)"
Cohesion: 0.13
Nodes (13): Velez — Features, `auth.yaml` — inter-service access control, `configuration.yaml` — app configuration, `dependencies.yaml` — external resources, Directory layout, How Velez processes the descriptor, `network.yaml` — ports and exposure, `resources.yaml` — hardware and placement (+5 more)

### Community 341 - "prepareVervConfig"
Cohesion: 0.31
Nodes (6): Context, CreateSmerd_Request, ServerStreamingServer, TaskStatus, WatchTask_Request, ServerStreamingClient

### Community 342 - "file_control_plane_api_proto_rawDescGZIP"
Cohesion: 0.36
Nodes (5): BoundResource, Context, Docker, newServiceResourcesStorage(), dockerServiceResourcesStorage

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

### Community 361 - "EnablePlugin_Response"
Cohesion: 0.09
Nodes (4): Message, CreateService_Response, GetService, RestartService_Response

### Community 365 - "fakeJobsStorage"
Cohesion: 0.29
Nodes (6): Context, ServiceBaseInfo, Step, Storage, UpsertServiceState(), upsertServiceState

### Community 367 - "Pattern"
Cohesion: 0.31
Nodes (8): extract_message(), main(), basicPostgresConstructor(), CreateRequest, Postgres, Postgres(), Constructor, Pattern

### Community 368 - "internal/clients (ns_list.go)"
Cohesion: 0.28
Nodes (6): getPgDbDsn, GetRgRootDsn(), Context, Docker, InspectResponse, Step

### Community 372 - ".GetServiceEnvironments"
Cohesion: 0.31
Nodes (6): MakoshBeAPIClient, NewClient(), NewServiceDiscovery(), HeaderOutgoingInterceptor(), ServiceDiscovery, UnaryClientInterceptor

### Community 373 - "connectServiceToVpnHandler"
Cohesion: 0.20
Nodes (5): ServiceDiscovery, CreateRequest, TailScaleContainerSidecar(), addMakoshRecordJob, connectServiceToVpnHandler

### Community 391 - "internal/storage (nodes.sql.go)"
Cohesion: 0.20
Nodes (4): Mutex, Service, fakeNodesStorage, fakeServicesStorage

### Community 429 - "Velez UI Redesign — Roadmap"
Cohesion: 0.22
Nodes (8): 1. Start the first node and let it self-host Headscale, 2. Point a second node at node A's Headscale server (manual — no join RPC exists), 3. Create a VCN namespace and attach a service to it (this part genuinely works), 4. Deploying a service "across the cluster" — current limits, Prerequisites, Quickstart: Multi-Node, Troubleshooting / gotchas, What's real today vs. what's a stub

### Community 430 - ".ListServices"
Cohesion: 0.36
Nodes (6): fromListServiceRequest(), Context, ListServices_Request, ListServices_Response, Impl, toListServiceResponse()

### Community 431 - "T30 — AppsPage + AppCard (logical services view)"
Cohesion: 0.17
Nodes (11): AppCard CSS highlights, AppCard data type, AppCard layout, AppCard props, AppsPage component, AppsPage CSS, Notes, Routes update (+3 more)

### Community 432 - "Container"
Cohesion: 0.22
Nodes (4): Pointer, Storage, Container, PluginsStorage

### Community 433 - ".Do"
Cohesion: 0.22
Nodes (5): createPgUserStep, CreatePgUserForNode(), Context, Step, initNodeStorageJob

### Community 434 - "Tasks"
Cohesion: 0.20
Nodes (9): 5.1 Loading states, 5.2 Error boundaries, 5.3 Empty states, 5.4 Toast / notification system, 5.5 Navigation & layout cleanup, Acceptance criteria, Goal, T5 — UX Polish (+1 more)

### Community 435 - "FileMountPoint"
Cohesion: 0.22
Nodes (8): 1. Start Velez, 2. Find your access key, 3. Create a service, 4. Create a deployment, 5. Verify it's running, Prerequisites, Quickstart: Single Node, Troubleshooting / gotchas

### Community 436 - ".enrichServiceWithSmerdData"
Cohesion: 0.18
Nodes (3): MessageState, GetHardware_Request, Version_Request

### Community 437 - "GetConfigNode_Response"
Cohesion: 0.18
Nodes (3): file_tasks_proto_init(), CopyToVolumeTaskPayload, init()

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

### Community 444 - "UpsertPluginParams"
Cohesion: 0.31
Nodes (6): Context, NullInt64, NullString, Queries, ListPluginsRow, UpsertPluginParams

### Community 446 - ".Do"
Cohesion: 0.54
Nodes (5): WithExposedPort(), WithInstanceName(), WithPassword(), WithPort(), Opt

### Community 448 - "Project Factory — Agent Context"
Cohesion: 0.25
Nodes (7): Agent Rules, Architecture, Code Style, Frontend, Go, Project Factory — Agent Context, Task System

### Community 449 - "Task 000 — Example Task Title"
Cohesion: 0.25
Nodes (7): Acceptance Criteria, Context, Do NOT change, Files to Create / Modify, Goal, Notes, Task 000 — Example Task Title

### Community 450 - "Task 001 — Health check endpoint"
Cohesion: 0.25
Nodes (7): Acceptance Criteria, Context, Do NOT change, Files to Create / Modify, Goal, Notes, Task 001 — Health check endpoint

### Community 451 - "renameContainerStep"
Cohesion: 0.38
Nodes (4): APIClient, Context, RenameContainer(), renameContainerStep

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

### Community 466 - ".ContainerCreate"
Cohesion: 0.38
Nodes (5): storeConfigStep, Configurator, Context, Step, StoreConfig()

### Community 467 - "dockerServiceDepsStorage"
Cohesion: 0.16
Nodes (12): NodeType, ServiceDependency, ServiceEnvironment, ServiceGraph, Time, Context, VervService, containerName() (+4 more)

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

### Community 482 - "GetService_Request"
Cohesion: 0.28
Nodes (6): Context, Storage, NewPgStateManager(), Context, SetupPgState(), pgState

### Community 483 - "db.go"
Cohesion: 0.38
Nodes (5): Queries, Queries, Tx, New(), DBTX

### Community 484 - "db.go"
Cohesion: 0.38
Nodes (5): Queries, Queries, Tx, New(), DBTX

### Community 485 - "prepareUpgradeVervConfigJob"
Cohesion: 0.38
Nodes (4): createNetworkIfMissing(), CreateSmerd_Request, createNetworkAPI, prepareUpgradeVervConfigJob

### Community 486 - "AddMakoshRecord"
Cohesion: 0.33
Nodes (5): AddMakoshRecord(), Context, ServiceDiscovery, Step, addMakoshRecord

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

### Community 495 - "ValuesMapping.tsx"
Cohesion: 0.40
Nodes (3): ActionButton(), ActionButtonProps, VolumesWidgetProps

### Community 496 - "InitMaster_Response"
Cohesion: 0.31
Nodes (3): TaskStatus, MD, fakeTaskStatusStream

### Community 497 - "fetch_by_api.go"
Cohesion: 0.40
Nodes (5): 2. PaaS platform, 3. UI redesign, 4. Genuinely open items carried forward from `docs/review/`, M1 — Core Platform, task-by-task, Velez Roadmap

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

### Community 507 - "SingleFunc"
Cohesion: 0.21
Nodes (7): Step, ValidateServiceName(), Context, SingleFunc(), RollbackableStep, singleFunc, Step

### Community 517 - "ListNodes"
Cohesion: 0.12
Nodes (3): Message, InitMaster_Request, ListVervPlugins_Request

### Community 539 - "Option B: shell factory with direct Ollama API calls"
Cohesion: 0.33
Nodes (5): Component, CSS, Props interface, T08 — SectionLabel, What it looks like

### Community 540 - "Config subscription is commented out (handleConfigurationSubscription)"
Cohesion: 0.33
Nodes (3): DropSmerd_Request, DropSmerd_Response_Error, DropSmerdTaskPayload

### Community 620 - "NewService"
Cohesion: 0.83
Nodes (3): Storage, NewService(), Service

### Community 682 - "pgStorage"
Cohesion: 0.53
Nodes (4): GetServiceReq, Context, Service, VervService

### Community 683 - "Configurator"
Cohesion: 0.50
Nodes (4): Configurator, MatreshkaBeAPI_SubscribeOnChangesClient, MatreshkaBeAPIClient, New()

### Community 689 - ".List"
Cohesion: 0.32
Nodes (6): ListServicesReq, Paging, TestEnrichServiceWithSmerdData(), TestServicePreservesLastDeployedAt(), TestServiceWithoutSmerd(), Optional

### Community 742 - "Jobs Migration — Open Questions"
Cohesion: 0.25
Nodes (7): ConnectServiceToVpn — resolved 2026-07-19, CopyToVolume (added during implementation), CopyToVolume — resolved 2026-07-19, Cross-cutting (all pipelines), EnableStatefullMode — resolved 2026-07-19, Jobs Migration — Open Questions, UpgradeSmerd — resolved 2026-07-19

### Community 753 - "Pipelines → Jobs Migration"
Cohesion: 0.40
Nodes (4): Migration checklist (repeat per pipeline), Pipelines → Jobs Migration, Status, Suggested pick-up order (easiest first)

### Community 757 - "VPN client key (Headscale auth key) storage in task context"
Cohesion: 0.50
Nodes (3): Decision, If this needs to change later, VPN client key (Headscale auth key) storage in task context

## Knowledge Gaps
- **1060 isolated node(s):** `status.sh script`, `go.vervstack.ru/Velez`, `UnsafeMatreshkaApiServer`, `UnsafeControlPlaneAPIServer`, `UnsafeServiceApiServer` (+1055 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **344 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `T` connect `pkg/web (index.ts)` to `Message`, `internal/clients (hardware_manager.go)`, `internal/transport (grpcServer)`, `pkg/web (ControlPlanePage.tsx)`, `internal/api (file_control_plane_api_proto_rawDescGZIP())`, `tests/e2e (suite_hello_world_cluster_test.go)`, `internal/config (launch.go)`, `internal/pipelines (create_and_inspect.go)`, `internal/pipelines (ConnectServiceToVpn())`, `runner`, `.List`, `internal/pipelines (do_copy_to_volume.go)`, `CreatePgUserForNode`, `pkg/web (verv_closed_network.pb.ts)`, `internal/domain (service.go)`, `tests/e2e (NewEnvironment())`, `internal/api (UnknownFields)`, `internal/api (GetHardware_Request)`, `internal/api (ListPlugins_Request)`?**
  _High betweenness centrality (0.186) - this node is a cross-community bridge._
- **Why does `file_velez_common_proto_init()` connect `internal/api (AssembleConfig_Response)` to `internal/api (velez_api.pb.go)`, `internal/api (Message)`, `internal/api (control_plane_api.pb.go)`, `GetConfigNode_Response`?**
  _High betweenness centrality (0.161) - this node is a cross-community bridge._
- **Why does `b64Decode()` connect `internal/pipelines (do_copy_to_volume.go)` to `pkg/web (index.ts)`?**
  _High betweenness centrality (0.136) - this node is a cross-community bridge._
- **Are the 59 inferred relationships involving `newFakeDocker()` (e.g. with `TestCreateScratchContainerJob_ContainerCreateError()` and `TestCreateScratchContainerJob_Rollback_NoContainerId_NoOp()`) actually correct?**
  _`newFakeDocker()` has 59 INFERRED edges - model-reasoned connections that need verification._
- **What connects `status.sh script`, `go.vervstack.ru/Velez`, `UnsafeMatreshkaApiServer` to the rest of the system?**
  _1074 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `internal/api (service_api.pb.gw.go)` be split into smaller, more focused modules?**
  _Cohesion score 0.058502799882110226 - nodes in this community are weakly interconnected._
- **Should `internal/api (velez_api.pb.gw.go)` be split into smaller, more focused modules?**
  _Cohesion score 0.06297029702970297 - nodes in this community are weakly interconnected._