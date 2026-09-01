# Graph Report - Velez-UI  (2026-09-01)

## Corpus Check
- 184 files · ~157,595 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1097 nodes · 2240 edges · 77 communities (69 shown, 8 thin omitted)
- Extraction: 99% EXTRACTED · 1% INFERRED · 0% AMBIGUOUS · INFERRED: 17 edges (avg confidence: 0.69)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `5023f1a3`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ServiceCard.tsx
- DeployWidget.tsx
- Router.tsx
- fetch.pb.ts
- StepsDialog.tsx
- ServicePageModel.ts
- service_api.pb.ts
- ServiceService
- control_plane_api.pb.ts
- velez_api.pb.ts
- compilerOptions
- HomePage.tsx
- devDependencies
- index.ts
- openStatefullPgDialog.tsx
- ServiceInfoPage.tsx
- service.ts
- velez_common.pb.ts
- Sidebar.tsx
- SmerdPage.tsx
- Architecture
- ControlPlaneService
- dependencies
- ControlPlanePage.tsx
- tasks.pb.ts
- velez.ts
- VelezAPI
- NodeCard.tsx
- VervClosedNetworkPage.tsx
- TopBar.tsx
- DeployRow.tsx
- useToaster
- StatefullPgSettingsScreen.tsx
- services.ts
- PluginMatrix.tsx
- NetworkTopologyMap.tsx
- DeployMenu.tsx
- scripts
- PluginManageDialog.tsx
- DeploymentStatusBadge.tsx
- compilerOptions
- VCNPeerTable.tsx
- ServiceGraph.tsx
- TasksApi
- api.ts
- knip.json
- ObservabilityTools.tsx
- package.json
- .fetchServiceMetrics
- React + TypeScript + Vite
- ListDeploymentsResponse
- eslint.config.js
- control_plane.ts
- Smerds.ts
- NodeCard.tsx
- useCredentialsStore
- DeploymentFilters.tsx
- .EnablePlugin
- ApiService
- VolumesWidget.tsx
- .GetService
- state.ts
- ValuesMapping.tsx
- StepsDialogHeader.tsx
- ObservabilityTools.tsx
- CreateSmerdRequest
- PortMapping.tsx
- .createRegistry
- .listPlugins
- UpdateEnvironmentResponse

## God Nodes (most connected - your core abstractions)
1. `useToaster` - 55 edges
2. `VervPlugin` - 35 edges
3. `useDialog` - 34 edges
4. `Button()` - 28 edges
5. `ServiceService` - 24 edges
6. `compilerOptions` - 21 edges
7. `VervPluginType` - 20 edges
8. `ControlPlaneService` - 20 edges
9. `NodeBaseInfo` - 18 edges
10. `ServiceApi` - 17 edges

## Surprising Connections (you probably didn't know these)
- `renderWithProviders()` --indirect_call--> `Dialog()`  [INFERRED]
  src/widgets/settings/EnvironmentsSettings/EnvironmentsSettings.test.tsx → src/app/hooks/dialog/Dialog.tsx
- `renderGate()` --indirect_call--> `AuthGate()`  [INFERRED]
  src/app/router/AuthGate.test.tsx → src/app/router/AuthGate.tsx
- `StatefullPgDeadPromptProps` --references--> `VervPlugin`  [EXTRACTED]
  src/dialogs/PluginManageDialog/plugins/StatefullPgDeadPrompt.tsx → src/model/services/VervPlugins.tsx
- `openStatefullPgDialog()` --indirect_call--> `StatefullPgSettingsScreen()`  [INFERRED]
  src/dialogs/PluginManageDialog/plugins/openStatefullPgDialog.tsx → src/dialogs/PluginManageDialog/plugins/screens/StatefullPgSettingsScreen.tsx
- `PluginContentProps` --references--> `VervPlugin`  [EXTRACTED]
  src/widgets/controlplane/PluginMatrix.tsx → src/model/services/VervPlugins.tsx

## Import Cycles
- None detected.

## Communities (77 total, 8 thin omitted)

### Community 0 - "ServiceCard.tsx"
Cohesion: 0.08
Nodes (32): ConnectionHealth, ConnectionHealthStatus, useConnectionHealth(), AppCardProps, EnvChip(), EnvChipProps, FreezeChip(), IncidentChip() (+24 more)

### Community 1 - "DeployWidget.tsx"
Cohesion: 0.20
Nodes (11): InfoMark(), InfoMarkProps, CreateSmerdReq, Port, Smerd, DeployPage(), DeployWidget(), DeployWidgetProps (+3 more)

### Community 2 - "Router.tsx"
Cohesion: 0.26
Nodes (8): Credentials, ls, GetInitReq(), InitReq, keyPath(), pathPrefixPath(), StoreApiKey(), StorePathPrefix()

### Community 3 - "fetch.pb.ts"
Cohesion: 0.05
Nodes (38): b64, b64Encode(), fetchStreamingRequest(), FlattenedRequestPayload, flattenRequestPayload(), getNewLineDelimitedJSONDecodingStream(), getNotifyEntityArrivalSink(), InitReq (+30 more)

### Community 4 - "StepsDialog.tsx"
Cohesion: 0.11
Nodes (12): FinalStepProps, Step, StepScreenProps, StepsDialogProps, TestContext, StepsDialogHeaderContent, TestContext, padIndex() (+4 more)

### Community 5 - "ServicePageModel.ts"
Cohesion: 0.07
Nodes (37): EnvCard(), EnvCardProps, ServiceEnvironment, ServiceGraphNode, ServiceResource, VervonomiconDocs, LIST_REQ, useGetServiceAboutQuery() (+29 more)

### Community 6 - "service_api.pb.ts"
Cohesion: 0.06
Nodes (34): AboutService, Absent, BaseCreateDeployRequest, BaseGetServiceResponse, BoundResource, CreateDeploy, CreateDeployRequestUpgrade, CreateDeployResponse (+26 more)

### Community 7 - "ServiceService"
Cohesion: 0.13
Nodes (4): ListDeploymentsResponse, ServiceApi, ServiceGraphData, ServiceService

### Community 8 - "control_plane_api.pb.ts"
Cohesion: 0.06
Nodes (33): Absent, BaseEnableHeadscaleServer, BaseEnablePluginRequest, ConnectSlave, ConnectSlaveRequest, ConnectSlaveResponse, CreateEnvironment, CreateRegistry (+25 more)

### Community 9 - "velez_api.pb.ts"
Cohesion: 0.08
Nodes (24): AssembleConfig, AssembleConfigRequest, AssembleConfigResponse, BreakConnections, BreakConnectionsRequest, BreakConnectionsResponse, CreateSmerd, DropSmerd (+16 more)

### Community 10 - "compilerOptions"
Cohesion: 0.08
Nodes (24): compilerOptions, allowImportingTsExtensions, allowJs, allowSyntheticDefaultImports, esModuleInterop, forceConsistentCasingInFileNames, isolatedModules, jsx (+16 more)

### Community 11 - "HomePage.tsx"
Cohesion: 0.15
Nodes (8): ServiceBaseInfo, formatLastDeployed(), HomePage(), ServiceCardActionsProps, ServicesSection(), ServicesSectionProps, SmerdsSectionProps, useListServicesQuery()

### Community 12 - "devDependencies"
Cohesion: 0.09
Nodes (23): devDependencies, eslint, eslint-import-resolver-typescript, @eslint/js, eslint-plugin-import-x, eslint-plugin-react, eslint-plugin-react-hooks, eslint-plugin-react-refresh (+15 more)

### Community 13 - "index.ts"
Cohesion: 0.14
Nodes (15): VervPluginState, VervPluginType, node, PluginManageDialogProps, metaByType, ServiceMeta, VervPlugin, nodes (+7 more)

### Community 14 - "openStatefullPgDialog.tsx"
Cohesion: 0.16
Nodes (8): Dialog(), DialogManager, useDialog, HeadscalePluginForm(), SimplePluginForm(), TestContext, StepsDialog(), AppsEmptyState()

### Community 15 - "ServiceInfoPage.tsx"
Cohesion: 0.14
Nodes (14): TagChip(), TagChipProps, ActionsRowProps, ServiceInfoPage(), ServicePageHeader(), ServiceTab, ServiceTagsStrip(), ServiceTagsStripProps (+6 more)

### Community 16 - "service.ts"
Cohesion: 0.13
Nodes (17): CreateDeployRequest, GetServiceEnvironmentsRequest, GetServiceGraphRequest, GetServiceMetricsRequest, GetServiceRequest, ListDeploymentsRequest, RemoveServiceRequest, RestartServiceRequest (+9 more)

### Community 17 - "velez_common.pb.ts"
Cohesion: 0.12
Nodes (16): ConfigFormat, Connection, Container, ContainerHardware, ContainerHealthcheck, ContainerSettings, FileConfig, Image (+8 more)

### Community 18 - "Sidebar.tsx"
Cohesion: 0.07
Nodes (21): useAuthGate(), AuthGate(), ping, renderGate(), MainLayout(), NAV_TO_ROUTE, NavId, ROUTE_TO_NAV (+13 more)

### Community 19 - "SmerdPage.tsx"
Cohesion: 0.16
Nodes (6): BreadcrumbsBarProps, Crumb, QueryErrorState(), QueryErrorStateProps, formatTimestamp(), SmerdMetaSection()

### Community 20 - "Architecture"
Cohesion: 0.13
Nodes (13): API calls, Architecture, Coding Rules, Commands, Dialogs, Environment, Exploration Rules, Layer structure (+5 more)

### Community 21 - "ControlPlaneService"
Cohesion: 0.12
Nodes (6): ControlPlaneAPI, CreateEnvironmentResponse, ListEnvironmentsResponse, ListRegistriesResponse, UpdateRegistryResponse, ControlPlaneService

### Community 22 - "dependencies"
Cohesion: 0.15
Nodes (13): dependencies, classnames, framer-motion, @microlink/react-json-view, react, react-dom, react-router-dom, react-tooltip (+5 more)

### Community 23 - "ControlPlanePage.tsx"
Cohesion: 0.23
Nodes (10): ListNodesResponse, ControlPlanePage(), NodeListProps, PluginsListProps, StatsGridProps, ListNodesQuery(), ListPluginsQuery(), NodesHealthStatus() (+2 more)

### Community 24 - "tasks.pb.ts"
Cohesion: 0.05
Nodes (45): AssembleConfigTaskPayload, ConnectServiceToVpnTaskPayload, CopyToVolumeTaskPayload, CreateServiceTaskPayload, CreateSmerdTaskPayload, DropSmerdTaskPayload, EnableStatefullTaskPayload, TasksApi (+37 more)

### Community 25 - "velez.ts"
Cohesion: 0.14
Nodes (12): GetHardwareResponse, ListSmerdsRequest, ListSmerdsResponse, SearchImagesRequest, SearchImagesResponse, toProto(), DeploySmerd(), DeploySmerdStream() (+4 more)

### Community 26 - "VelezAPI"
Cohesion: 0.18
Nodes (3): VelezAPI, VersionResponse, ListImages()

### Community 27 - "NodeCard.tsx"
Cohesion: 0.15
Nodes (14): Environment, CreateServiceRequest, EnvironmentStore, NOTE: this is unrelated to `ServiceEnvironment`, useEnvironmentStore, EnvironmentDeleteDialog(), EnvironmentDeleteDialogProps, EnvironmentManageDialog() (+6 more)

### Community 28 - "VervClosedNetworkPage.tsx"
Cohesion: 0.07
Nodes (28): DeploymentStatus, Badge(), BadgeProps, Level, StatCard(), StatCardProps, CodeBlock(), CodeBlockProps (+20 more)

### Community 29 - "TopBar.tsx"
Cohesion: 0.27
Nodes (9): EnvironmentCreateDialog(), EnvironmentCreateDialogProps, CreateEnvironmentMutation(), ENVIRONMENTS_QUERY_KEY, IsStatefullModeEnabled(), ListEnvironmentsQuery(), REGISTRIES_QUERY_KEY, EnvironmentsSettings() (+1 more)

### Community 30 - "DeployRow.tsx"
Cohesion: 0.31
Nodes (8): DeploymentInfo, DeploymentHistoryProps, DeployRow(), DeployRowProps, formatTimestamp(), getStatusClass(), parseImageTag(), TODO: replace with git commit hash when GetServiceGraph adds git field

### Community 31 - "useToaster"
Cohesion: 0.31
Nodes (8): Toast, Toaster, useToaster, RemoveServiceDialog(), RemoveServiceDialogProps, ServiceCardActions(), Toast(), Toaster()

### Community 32 - "StatefullPgSettingsScreen.tsx"
Cohesion: 0.17
Nodes (11): Registry, Checkbox(), CheckboxProps, ChoiceProps, RegistryDeleteDialog(), RegistryDeleteDialogProps, RegistryConnectionFieldsProps, RegistryManageDialog() (+3 more)

### Community 33 - "services.ts"
Cohesion: 0.29
Nodes (5): SkeletonLoader(), SkeletonLoaderProps, SkeletonNodeCard(), SkeletonServiceCard(), SkeletonSmerdRow()

### Community 34 - "PluginMatrix.tsx"
Cohesion: 0.21
Nodes (11): getStatusRank(), mapVervPluginStateToPluginStatus(), NodeHeader(), PluginContent(), PluginContentProps, PluginMatrix(), PluginMatrixProps, sortPluginsByStatus() (+3 more)

### Community 35 - "NetworkTopologyMap.tsx"
Cohesion: 0.29
Nodes (4): SectionLabel(), SectionLabelProps, NodeHealthListProps, NodeProps

### Community 36 - "DeployMenu.tsx"
Cohesion: 0.31
Nodes (7): SkeletonDeploymentHistory(), DeployMenu(), DeployMenuProps, TabsOptions, TERMINAL_STATUSES, ActionsRow(), ListDeploymentsByServiceNameQuery()

### Community 37 - "scripts"
Cohesion: 0.22
Nodes (9): scripts, build, dev, gen, knip, lint, lint:css, lint:js (+1 more)

### Community 38 - "PluginManageDialog.tsx"
Cohesion: 0.27
Nodes (7): VervPlugin, Routes, formatState(), mapStateToStatus(), pluginForms, PluginManageDialog(), UnknownPlugin()

### Community 39 - "DeploymentStatusBadge.tsx"
Cohesion: 0.16
Nodes (13): NodeBaseInfo, NodeStatus, NodesPanel(), NodesPanelProps, plugins, findMasterNodeId(), layoutOuterNodes(), NodeTopologyGraph() (+5 more)

### Community 40 - "compilerOptions"
Cohesion: 0.22
Nodes (8): compilerOptions, allowSyntheticDefaultImports, composite, module, moduleResolution, skipLibCheck, strict, include

### Community 41 - "VCNPeerTable.tsx"
Cohesion: 0.21
Nodes (6): AppCard(), Button(), ButtonProps, CreateAppDialog(), deriveAppName(), StepsDialogFooterProps

### Community 42 - "ServiceGraph.tsx"
Cohesion: 0.19
Nodes (11): RegistryType, Props, RegistryImagePicker(), RegistryConnectionFieldsProps, RegistryCreateDialog(), RegistryCreateDialogProps, CreateRegistryMutation(), ListRegistriesQuery() (+3 more)

### Community 43 - "TasksApi"
Cohesion: 0.47
Nodes (5): SmerdPage(), FetchSmerd(), FetchSmerdsByServiceId(), ListSmerdsByServiceIdQuery(), useGetSmerdQuery()

### Community 45 - "knip.json"
Cohesion: 0.33
Nodes (5): entry, ignore, project, $schema, knip

### Community 46 - "ObservabilityTools.tsx"
Cohesion: 0.22
Nodes (12): Smerd, SmerdStatus, AppsPage(), DeploymentsPage(), ViewMode, SearchPage(), LOCAL_NODE, mapSmerdStatus() (+4 more)

### Community 47 - "package.json"
Cohesion: 0.40
Nodes (4): name, private, trustedDependencies, type

### Community 50 - "ListDeploymentsResponse"
Cohesion: 0.15
Nodes (6): InputProps, StyleProps, KeyValueProps, PlainMapProps, SearchParam, SearchProps

### Community 59 - "control_plane.ts"
Cohesion: 0.22
Nodes (8): CreateEnvironmentRequest, CreateRegistryRequest, DeleteEnvironmentRequest, DeleteRegistryRequest, EnablePluginRequest, ListPluginsRequest, UpdateEnvironmentRequest, UpdateRegistryRequest

### Community 60 - "Smerds.ts"
Cohesion: 0.33
Nodes (8): Port, Volume, Bind, fromProto(), fromProtoPort(), fromProtoPorts(), fromProtoVolume(), fromProtoVolumes()

### Community 61 - "NodeCard.tsx"
Cohesion: 0.33
Nodes (7): IconButton(), IconButtonProps, isMetricAmber(), isMetricRed(), mapNodeStatus(), NodeCard(), NodeCardProps

### Community 62 - "useCredentialsStore"
Cohesion: 0.32
Nodes (6): AUTH_GATE_QUERY_KEY, AuthGateStatus, useCredentialsStore, LoginPage(), NewServicePage(), VelezService

### Community 63 - "DeploymentFilters.tsx"
Cohesion: 0.25
Nodes (4): DeploymentFiltersProps, ENV_OPTIONS, STATUS_OPTIONS, ViewMode

### Community 64 - ".EnablePlugin"
Cohesion: 0.29
Nodes (3): EnableHeadscaleServer, EnablePluginResponse, EnableStatefullCluster

### Community 65 - "ApiService"
Cohesion: 0.57
Nodes (3): InitReq, ApiService, withRetries()

### Community 66 - "VolumesWidget.tsx"
Cohesion: 0.38
Nodes (5): ValuesPair(), VolumeMappingProps, Volume, VolumesWidget(), VolumesWidgetProps

### Community 68 - "state.ts"
Cohesion: 0.47
Nodes (4): Settings, SettingsBase, storeToLocalStorage(), useSettings()

### Community 69 - "ValuesMapping.tsx"
Cohesion: 0.40
Nodes (3): ActionButton(), ActionButtonProps, VolumesWidgetProps

### Community 70 - "StepsDialogHeader.tsx"
Cohesion: 0.40
Nodes (3): StepsDialogHeader(), StepsDialogHeaderProps, warnFieldMissing()

### Community 71 - "ObservabilityTools.tsx"
Cohesion: 0.33
Nodes (4): ObservabilityToolsProps, TODO: navigate to tool, TODO: add observability support in Velez, Tool

### Community 72 - "CreateSmerdRequest"
Cohesion: 0.67
Nodes (3): CreateSmerdRequest, DeploymentWidget(), DeployWidgetProps

### Community 73 - "PortMapping.tsx"
Cohesion: 0.67
Nodes (3): PortMapping(), PortMappingProps, getLinkToPort()

## Knowledge Gaps
- **368 isolated node(s):** `localPlugin`, `$schema`, `entry`, `project`, `ignore` (+363 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **8 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `useToaster` connect `useToaster` to `StatefullPgSettingsScreen.tsx`, `ServiceCard.tsx`, `DeployWidget.tsx`, `DeployMenu.tsx`, `ServicePageModel.ts`, `PluginManageDialog.tsx`, `VCNPeerTable.tsx`, `ServiceGraph.tsx`, `HomePage.tsx`, `TasksApi`, `openStatefullPgDialog.tsx`, `ObservabilityTools.tsx`, `ServiceInfoPage.tsx`, `SmerdPage.tsx`, `tasks.pb.ts`, `NodeCard.tsx`, `TopBar.tsx`?**
  _High betweenness centrality (0.059) - this node is a cross-community bridge._
- **Why does `ServiceApi` connect `ServiceService` to `.GetService`, `PluginManageDialog.tsx`, `service_api.pb.ts`, `api.ts`, `.fetchServiceMetrics`, `service.ts`, `NodeCard.tsx`?**
  _High betweenness centrality (0.028) - this node is a cross-community bridge._
- **Why does `useDialog` connect `openStatefullPgDialog.tsx` to `StatefullPgSettingsScreen.tsx`, `PluginMatrix.tsx`, `StepsDialog.tsx`, `DeployMenu.tsx`, `PluginManageDialog.tsx`, `VCNPeerTable.tsx`, `ServiceGraph.tsx`, `HomePage.tsx`, `index.ts`, `ObservabilityTools.tsx`, `ServiceInfoPage.tsx`, `tasks.pb.ts`, `NodeCard.tsx`, `TopBar.tsx`?**
  _High betweenness centrality (0.028) - this node is a cross-community bridge._
- **What connects `localPlugin`, `$schema`, `entry` to the rest of the system?**
  _374 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ServiceCard.tsx` be split into smaller, more focused modules?**
  _Cohesion score 0.07993197278911565 - nodes in this community are weakly interconnected._
- **Should `fetch.pb.ts` be split into smaller, more focused modules?**
  _Cohesion score 0.04591836734693878 - nodes in this community are weakly interconnected._
- **Should `StepsDialog.tsx` be split into smaller, more focused modules?**
  _Cohesion score 0.11083743842364532 - nodes in this community are weakly interconnected._