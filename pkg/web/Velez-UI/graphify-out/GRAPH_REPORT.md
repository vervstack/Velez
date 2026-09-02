# Graph Report - Velez-UI  (2026-09-02)

## Corpus Check
- 184 files · ~157,603 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1097 nodes · 2242 edges · 71 communities (61 shown, 10 thin omitted)
- Extraction: 99% EXTRACTED · 1% INFERRED · 0% AMBIGUOUS · INFERRED: 17 edges (avg confidence: 0.69)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `5bdda42f`
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
- ObservabilityTools.tsx
- .createRegistry

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
- `StepsDialog()` --calls--> `useDialog`  [EXTRACTED]
  src/dialogs/StepsDialog/StepsDialog.tsx → src/app/hooks/dialog/Dialog.tsx
- `renderWithProviders()` --indirect_call--> `Dialog()`  [INFERRED]
  src/widgets/settings/EnvironmentsSettings/EnvironmentsSettings.test.tsx → src/app/hooks/dialog/Dialog.tsx
- `ServiceCardActions()` --calls--> `useToaster`  [EXTRACTED]
  src/pages/home/HomePage.tsx → src/app/hooks/toaster/Toaster.ts
- `renderGate()` --indirect_call--> `AuthGate()`  [INFERRED]
  src/app/router/AuthGate.test.tsx → src/app/router/AuthGate.tsx
- `StatefullPgDeadPromptProps` --references--> `VervPlugin`  [EXTRACTED]
  src/dialogs/PluginManageDialog/plugins/StatefullPgDeadPrompt.tsx → src/model/services/VervPlugins.tsx

## Import Cycles
- None detected.

## Communities (71 total, 10 thin omitted)

### Community 0 - "ServiceCard.tsx"
Cohesion: 0.06
Nodes (42): Smerd, SmerdStatus, AppCardProps, EnvChip(), EnvChipProps, FreezeChip(), IncidentChip(), MiniBar() (+34 more)

### Community 1 - "DeployWidget.tsx"
Cohesion: 0.05
Nodes (39): CreateSmerdRequest, Port, Volume, ActionButton(), ActionButtonProps, InfoMark(), InfoMarkProps, InputProps (+31 more)

### Community 2 - "Router.tsx"
Cohesion: 0.26
Nodes (8): Credentials, ls, GetInitReq(), InitReq, keyPath(), pathPrefixPath(), StoreApiKey(), StorePathPrefix()

### Community 3 - "fetch.pb.ts"
Cohesion: 0.05
Nodes (38): b64, b64Encode(), fetchStreamingRequest(), FlattenedRequestPayload, flattenRequestPayload(), getNewLineDelimitedJSONDecodingStream(), getNotifyEntityArrivalSink(), InitReq (+30 more)

### Community 4 - "StepsDialog.tsx"
Cohesion: 0.11
Nodes (13): FinalStepProps, Step, StepsDialog(), StepsDialogProps, StepsDialogHeader(), StepsDialogHeaderContent, StepsDialogHeaderProps, warnFieldMissing() (+5 more)

### Community 5 - "ServicePageModel.ts"
Cohesion: 0.23
Nodes (6): ServiceResource, useGetServiceResourcesQuery(), ResourceCard(), ResourceCardProps, ResourcesSection(), ResourcesSectionProps

### Community 6 - "service_api.pb.ts"
Cohesion: 0.06
Nodes (34): AboutService, Absent, BaseCreateDeployRequest, BaseGetServiceResponse, BoundResource, CreateDeploy, CreateDeployRequestUpgrade, CreateDeployResponse (+26 more)

### Community 7 - "ServiceService"
Cohesion: 0.11
Nodes (6): ListDeploymentsResponse, ServiceApi, VervAppService, ServiceAbout, ServiceGraphData, ServiceService

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
Cohesion: 0.12
Nodes (11): ServiceBaseInfo, QueryErrorState(), QueryErrorStateProps, formatLastDeployed(), HomePage(), ServiceCardActions(), ServiceCardActionsProps, ServicesSection() (+3 more)

### Community 12 - "devDependencies"
Cohesion: 0.09
Nodes (23): devDependencies, eslint, eslint-import-resolver-typescript, @eslint/js, eslint-plugin-import-x, eslint-plugin-react, eslint-plugin-react-hooks, eslint-plugin-react-refresh (+15 more)

### Community 13 - "index.ts"
Cohesion: 0.22
Nodes (9): VervPluginState, VervPluginType, Routes, node, PluginManageDialogProps, metaByType, ServiceMeta, VervPlugin (+1 more)

### Community 14 - "openStatefullPgDialog.tsx"
Cohesion: 0.18
Nodes (4): TestContext, StepScreenProps, TestContext, TestContext

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
Cohesion: 0.05
Nodes (36): CreateServiceRequest, AUTH_GATE_QUERY_KEY, AuthGateStatus, useAuthGate(), ConnectionHealth, ConnectionHealthStatus, useConnectionHealth(), AuthGate() (+28 more)

### Community 19 - "SmerdPage.tsx"
Cohesion: 0.22
Nodes (7): formatTimestamp(), SmerdMetaSection(), SmerdPage(), FetchSmerd(), FetchSmerdsByServiceId(), ListSmerdsByServiceIdQuery(), useGetSmerdQuery()

### Community 20 - "Architecture"
Cohesion: 0.13
Nodes (13): API calls, Architecture, Coding Rules, Commands, Dialogs, Environment, Exploration Rules, Layer structure (+5 more)

### Community 21 - "ControlPlaneService"
Cohesion: 0.12
Nodes (6): ControlPlaneAPI, CreateEnvironmentResponse, ListEnvironmentsResponse, UpdateEnvironmentResponse, ControlPlaneService, toServices()

### Community 22 - "dependencies"
Cohesion: 0.15
Nodes (13): dependencies, classnames, framer-motion, @microlink/react-json-view, react, react-dom, react-router-dom, react-tooltip (+5 more)

### Community 23 - "ControlPlanePage.tsx"
Cohesion: 0.23
Nodes (10): ListNodesResponse, ControlPlanePage(), NodeListProps, PluginsListProps, StatsGridProps, ListNodesQuery(), ListPluginsQuery(), NodesHealthStatus() (+2 more)

### Community 24 - "tasks.pb.ts"
Cohesion: 0.06
Nodes (43): AssembleConfigTaskPayload, ConnectServiceToVpnTaskPayload, CopyToVolumeTaskPayload, CreateServiceTaskPayload, CreateSmerdTaskPayload, DropSmerdTaskPayload, EnableStatefullTaskPayload, TasksApi (+35 more)

### Community 25 - "velez.ts"
Cohesion: 0.19
Nodes (11): GetHardwareResponse, ListSmerdsRequest, ListSmerdsResponse, SearchImagesRequest, SearchImagesResponse, toProto(), DeploySmerd(), DeploySmerdStream() (+3 more)

### Community 26 - "VelezAPI"
Cohesion: 0.17
Nodes (4): VelezAPI, VersionResponse, ListImages(), VelezService

### Community 27 - "NodeCard.tsx"
Cohesion: 0.21
Nodes (10): Environment, EnvironmentStore, NOTE: this is unrelated to `ServiceEnvironment`, useEnvironmentStore, EnvironmentDeleteDialog(), EnvironmentDeleteDialogProps, EnvironmentManageDialog(), EnvironmentManageDialogProps (+2 more)

### Community 28 - "VervClosedNetworkPage.tsx"
Cohesion: 0.07
Nodes (28): DeploymentStatus, Badge(), BadgeProps, Level, StatCard(), StatCardProps, CodeBlock(), CodeBlockProps (+20 more)

### Community 29 - "TopBar.tsx"
Cohesion: 0.36
Nodes (7): EnvironmentCreateDialog(), EnvironmentCreateDialogProps, CreateEnvironmentMutation(), IsStatefullModeEnabled(), ListEnvironmentsQuery(), EnvironmentsSettings(), EnvRow()

### Community 30 - "DeployRow.tsx"
Cohesion: 0.31
Nodes (8): DeploymentInfo, DeploymentHistoryProps, DeployRow(), DeployRowProps, formatTimestamp(), getStatusClass(), parseImageTag(), TODO: replace with git commit hash when GetServiceGraph adds git field

### Community 31 - "useToaster"
Cohesion: 0.27
Nodes (9): Toast, Toaster, useToaster, StatefullPgDeadPrompt(), StatefullPgDeadPromptProps, RemoveServiceDialog(), RemoveServiceDialogProps, Toast() (+1 more)

### Community 32 - "StatefullPgSettingsScreen.tsx"
Cohesion: 0.10
Nodes (24): Registry, RegistryType, Checkbox(), CheckboxProps, ChoiceProps, Props, RegistryImagePicker(), RegistryConnectionFieldsProps (+16 more)

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
Cohesion: 0.36
Nodes (6): VervPlugin, formatState(), mapStateToStatus(), pluginForms, PluginManageDialog(), UnknownPlugin()

### Community 39 - "DeploymentStatusBadge.tsx"
Cohesion: 0.16
Nodes (13): NodeBaseInfo, NodeStatus, NodesPanel(), NodesPanelProps, plugins, findMasterNodeId(), layoutOuterNodes(), NodeTopologyGraph() (+5 more)

### Community 40 - "compilerOptions"
Cohesion: 0.22
Nodes (8): compilerOptions, allowSyntheticDefaultImports, composite, module, moduleResolution, skipLibCheck, strict, include

### Community 41 - "VCNPeerTable.tsx"
Cohesion: 0.16
Nodes (11): Dialog(), DialogManager, useDialog, AppCard(), Button(), ButtonProps, CreateAppDialog(), deriveAppName() (+3 more)

### Community 42 - "ServiceGraph.tsx"
Cohesion: 0.29
Nodes (9): LIST_REQ, useGetServiceAboutQuery(), useGetServiceMetricsQuery(), MetricTile(), MetricTileProps, progressBarFillClass(), ServiceHero(), ServiceHeroProps (+1 more)

### Community 43 - "TasksApi"
Cohesion: 0.33
Nodes (7): EnvCard(), EnvCardProps, ServiceEnvironment, ServiceGraphNode, useListServiceEnvsQuery(), EnvSwitcher(), EnvSwitcherProps

### Community 45 - "knip.json"
Cohesion: 0.33
Nodes (5): entry, ignore, project, $schema, knip

### Community 46 - "ObservabilityTools.tsx"
Cohesion: 0.27
Nodes (8): VervonomiconDocs, useGetVervonomiconQuery(), highlightLines(), isAllEmpty(), TabKey, TABS, Vervonomicon(), VervonomiconProps

### Community 47 - "package.json"
Cohesion: 0.40
Nodes (4): name, private, trustedDependencies, type

### Community 50 - "ListDeploymentsResponse"
Cohesion: 0.46
Nodes (7): useGetServiceGraphQuery(), labelLines(), nodeEdgeX(), renderNode(), ServiceGraph(), ServiceGraphProps, spreadY()

### Community 59 - "control_plane.ts"
Cohesion: 0.22
Nodes (8): CreateEnvironmentRequest, CreateRegistryRequest, DeleteEnvironmentRequest, DeleteRegistryRequest, EnablePluginRequest, ListPluginsRequest, UpdateEnvironmentRequest, UpdateRegistryRequest

### Community 60 - "Smerds.ts"
Cohesion: 0.25
Nodes (7): createEnvironment, deleteEnvironment, environments, listEnvironments, renderWithProviders(), statefullPlugin, updateEnvironment

### Community 61 - "NodeCard.tsx"
Cohesion: 0.33
Nodes (7): IconButton(), IconButtonProps, isMetricAmber(), isMetricRed(), mapNodeStatus(), NodeCard(), NodeCardProps

### Community 64 - ".EnablePlugin"
Cohesion: 0.29
Nodes (3): EnableHeadscaleServer, EnablePluginResponse, EnableStatefullCluster

### Community 65 - "ApiService"
Cohesion: 0.27
Nodes (7): InitReq, Settings, SettingsBase, storeToLocalStorage(), useSettings(), ApiService, withRetries()

### Community 71 - "ObservabilityTools.tsx"
Cohesion: 0.33
Nodes (4): ObservabilityToolsProps, TODO: navigate to tool, TODO: add observability support in Velez, Tool

## Knowledge Gaps
- **368 isolated node(s):** `localPlugin`, `$schema`, `entry`, `project`, `ignore` (+363 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **10 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `useToaster` connect `useToaster` to `StatefullPgSettingsScreen.tsx`, `ServiceCard.tsx`, `DeployWidget.tsx`, `DeployMenu.tsx`, `VCNPeerTable.tsx`, `ServiceGraph.tsx`, `HomePage.tsx`, `index.ts`, `ServiceInfoPage.tsx`, `ListDeploymentsResponse`, `SmerdPage.tsx`, `Sidebar.tsx`, `tasks.pb.ts`, `NodeCard.tsx`, `TopBar.tsx`?**
  _High betweenness centrality (0.059) - this node is a cross-community bridge._
- **Why does `ServiceApi` connect `ServiceService` to `service_api.pb.ts`, `api.ts`, `index.ts`, `.fetchServiceMetrics`, `service.ts`, `Sidebar.tsx`, `NodeCard.tsx`?**
  _High betweenness centrality (0.028) - this node is a cross-community bridge._
- **Why does `useDialog` connect `VCNPeerTable.tsx` to `StatefullPgSettingsScreen.tsx`, `ServiceCard.tsx`, `PluginMatrix.tsx`, `StepsDialog.tsx`, `DeployMenu.tsx`, `HomePage.tsx`, `index.ts`, `openStatefullPgDialog.tsx`, `ServiceInfoPage.tsx`, `tasks.pb.ts`, `NodeCard.tsx`, `TopBar.tsx`?**
  _High betweenness centrality (0.028) - this node is a cross-community bridge._
- **What connects `localPlugin`, `$schema`, `entry` to the rest of the system?**
  _374 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ServiceCard.tsx` be split into smaller, more focused modules?**
  _Cohesion score 0.061955965181771634 - nodes in this community are weakly interconnected._
- **Should `DeployWidget.tsx` be split into smaller, more focused modules?**
  _Cohesion score 0.053185271770894216 - nodes in this community are weakly interconnected._
- **Should `fetch.pb.ts` be split into smaller, more focused modules?**
  _Cohesion score 0.04591836734693878 - nodes in this community are weakly interconnected._