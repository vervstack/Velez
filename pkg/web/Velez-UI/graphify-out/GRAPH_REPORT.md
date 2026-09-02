# Graph Report - Velez-UI  (2026-09-02)

## Corpus Check
- 197 files · ~158,780 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1137 nodes · 2345 edges · 81 communities (73 shown, 8 thin omitted)
- Extraction: 99% EXTRACTED · 1% INFERRED · 0% AMBIGUOUS · INFERRED: 19 edges (avg confidence: 0.66)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `d7191a03`
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
- CreateAppDialog.tsx
- ServiceLabelBadge.tsx
- ObservabilityTools.tsx
- PortsWidget.tsx
- VCNPeerTable.tsx
- .createRegistry
- NewServicePage.tsx
- VolumesWidget.tsx
- NetworkTopologyMap.tsx
- DeploymentStatusBadge.tsx
- ValuesMapping.tsx
- .fetchServiceResources

## God Nodes (most connected - your core abstractions)
1. `useToaster` - 56 edges
2. `VervPlugin` - 35 edges
3. `useDialog` - 32 edges
4. `Button()` - 27 edges
5. `ServiceService` - 24 edges
6. `compilerOptions` - 21 edges
7. `VervPluginType` - 20 edges
8. `ControlPlaneService` - 20 edges
9. `NodeBaseInfo` - 18 edges
10. `ServiceApi` - 17 edges

## Surprising Connections (you probably didn't know these)
- `openStatefullPgDialog()` --indirect_call--> `StatefullPgSettingsScreen()`  [INFERRED]
  src/dialogs/PluginManageDialog/plugins/openStatefullPgDialog.tsx → src/dialogs/PluginManageDialog/plugins/screens/StatefullPgSettingsScreen.tsx
- `StepsDialog()` --calls--> `useDialog`  [EXTRACTED]
  src/dialogs/StepsDialog/StepsDialog.tsx → src/app/hooks/dialog/Dialog.tsx
- `renderWithProviders()` --indirect_call--> `Dialog()`  [INFERRED]
  src/widgets/settings/EnvironmentsSettings/EnvironmentsSettings.test.tsx → src/app/hooks/dialog/Dialog.tsx
- `ServiceCardActions()` --calls--> `useToaster`  [EXTRACTED]
  src/pages/home/HomePage.tsx → src/app/hooks/toaster/Toaster.ts
- `renderGate()` --indirect_call--> `AuthGate()`  [INFERRED]
  src/app/router/AuthGate.test.tsx → src/app/router/AuthGate.tsx

## Import Cycles
- None detected.

## Communities (81 total, 8 thin omitted)

### Community 0 - "ServiceCard.tsx"
Cohesion: 0.06
Nodes (46): ServiceBaseInfo, SmerdStatus, EnvChip(), EnvChipProps, FreezeChip(), IncidentChip(), MiniBar(), MiniBarProps (+38 more)

### Community 1 - "DeployWidget.tsx"
Cohesion: 0.15
Nodes (6): InputProps, StyleProps, KeyValueProps, PlainMapProps, SearchParam, SearchProps

### Community 2 - "Router.tsx"
Cohesion: 0.26
Nodes (8): Credentials, ls, GetInitReq(), InitReq, keyPath(), pathPrefixPath(), StoreApiKey(), StorePathPrefix()

### Community 3 - "fetch.pb.ts"
Cohesion: 0.05
Nodes (38): b64, b64Encode(), fetchStreamingRequest(), FlattenedRequestPayload, flattenRequestPayload(), getNewLineDelimitedJSONDecodingStream(), getNotifyEntityArrivalSink(), InitReq (+30 more)

### Community 4 - "StepsDialog.tsx"
Cohesion: 0.08
Nodes (18): OpenWizardDialog(), TestContext, FinalStepProps, Step, StepScreenProps, StepsDialog(), StepsDialogProps, TestContext (+10 more)

### Community 5 - "ServicePageModel.ts"
Cohesion: 0.06
Nodes (39): EnvCard(), EnvCardProps, ServiceAbout, ServiceEnvironment, ServiceGraphNode, ServiceResource, VervonomiconDocs, LIST_REQ (+31 more)

### Community 6 - "service_api.pb.ts"
Cohesion: 0.06
Nodes (32): AboutService, Absent, BaseCreateDeployRequest, BaseGetServiceResponse, CreateDeploy, CreateDeployRequestUpgrade, CreateDeployResponse, CreateService (+24 more)

### Community 7 - "ServiceService"
Cohesion: 0.11
Nodes (5): ListDeploymentsResponse, ServiceApi, VervAppService, ServiceGraphData, ServiceService

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
Nodes (8): formatLastDeployed(), HomePage(), readIncludeInternal(), ServiceCardActions(), ServiceCardActionsProps, ServicesSection(), ServicesSectionProps, SmerdsSectionProps

### Community 12 - "devDependencies"
Cohesion: 0.09
Nodes (23): devDependencies, eslint, eslint-import-resolver-typescript, @eslint/js, eslint-plugin-import-x, eslint-plugin-react, eslint-plugin-react-hooks, eslint-plugin-react-refresh (+15 more)

### Community 13 - "index.ts"
Cohesion: 0.16
Nodes (12): VervPluginState, VervPluginType, node, PluginManageDialogProps, StatefullPgDeadPromptProps, metaByType, ServiceMeta, VervPlugin (+4 more)

### Community 14 - "openStatefullPgDialog.tsx"
Cohesion: 0.29
Nodes (5): queryClient, openStatefullPgDialog(), StatefullPgOverviewScreen(), StatefullPgReviewScreen(), PipelineFlowLoader()

### Community 15 - "ServiceInfoPage.tsx"
Cohesion: 0.26
Nodes (9): TagChip(), TagChipProps, Props, ServiceTagsStrip(), FetchSmerd(), FetchSmerdsByServiceId(), getSmerdTags(), ListSmerdsByServiceIdQuery() (+1 more)

### Community 16 - "service.ts"
Cohesion: 0.12
Nodes (19): BoundResource, CreateDeployRequest, GetServiceEnvironmentsRequest, GetServiceGraphRequest, GetServiceMetricsRequest, GetServiceRequest, GetServiceResourcesRequest, ListDeploymentsRequest (+11 more)

### Community 17 - "velez_common.pb.ts"
Cohesion: 0.12
Nodes (16): ConfigFormat, Connection, Container, ContainerHardware, ContainerHealthcheck, ContainerSettings, FileConfig, Image (+8 more)

### Community 18 - "Sidebar.tsx"
Cohesion: 0.05
Nodes (33): AUTH_GATE_QUERY_KEY, AuthGateStatus, useAuthGate(), ConnectionHealth, ConnectionHealthStatus, useConnectionHealth(), AuthGate(), ping (+25 more)

### Community 19 - "SmerdPage.tsx"
Cohesion: 0.20
Nodes (7): Smerd, useBreadcrumbs, ServiceDetailLayout(), formatTimestamp(), SmerdMetaSection(), SmerdPage(), LeftZone()

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
Nodes (36): AssembleConfigTaskPayload, ConnectServiceToVpnTaskPayload, CopyToVolumeTaskPayload, CreateServiceTaskPayload, CreateSmerdTaskPayload, DropSmerdTaskPayload, EnableStatefullTaskPayload, TasksApi (+28 more)

### Community 25 - "velez.ts"
Cohesion: 0.21
Nodes (11): GetHardwareResponse, ListSmerdsRequest, ListSmerdsResponse, SearchImagesRequest, SearchImagesResponse, toProto(), DeploySmerd(), DeploySmerdStream() (+3 more)

### Community 26 - "VelezAPI"
Cohesion: 0.15
Nodes (4): VelezAPI, VersionResponse, ListImages(), VelezService

### Community 27 - "NodeCard.tsx"
Cohesion: 0.31
Nodes (6): Environment, EnvironmentStore, NOTE: this is unrelated to `ServiceEnvironment`, useEnvironmentStore, EnvironmentDeleteDialogProps, EnvironmentManageDialogProps

### Community 28 - "VervClosedNetworkPage.tsx"
Cohesion: 0.21
Nodes (10): Level, StatCard(), StatCardProps, CodeBlock(), CodeBlockProps, tokenizeLine(), MOCK_EDGES, MOCK_NODES (+2 more)

### Community 29 - "TopBar.tsx"
Cohesion: 0.31
Nodes (8): EnvironmentCreateDialog(), EnvironmentCreateDialogProps, CreateEnvironmentMutation(), IsStatefullModeEnabled(), ListEnvironmentsQuery(), EnvironmentsSettings(), EnvRow(), renderWithProviders()

### Community 30 - "DeployRow.tsx"
Cohesion: 0.31
Nodes (8): DeploymentInfo, DeploymentHistoryProps, DeployRow(), DeployRowProps, formatTimestamp(), getStatusClass(), parseImageTag(), TODO: replace with git commit hash when GetServiceGraph adds git field

### Community 31 - "useToaster"
Cohesion: 0.22
Nodes (13): Toast, Toaster, useToaster, EnvironmentDeleteDialog(), EnvironmentManageDialog(), RemoveServiceDialog(), RemoveServiceDialogProps, Props (+5 more)

### Community 32 - "StatefullPgSettingsScreen.tsx"
Cohesion: 0.12
Nodes (21): Registry, RegistryType, Checkbox(), CheckboxProps, RegistryConnectionFieldsProps, RegistryCreateDialog(), RegistryCreateDialogProps, RegistryDeleteDialog() (+13 more)

### Community 33 - "services.ts"
Cohesion: 0.18
Nodes (8): SkeletonLoader(), SkeletonLoaderProps, SkeletonNodeCard(), SkeletonServiceCard(), SkeletonSmerdRow(), ServiceInfoPage(), ServiceRouteDispatch(), VervCoreServicePage()

### Community 34 - "PluginMatrix.tsx"
Cohesion: 0.36
Nodes (8): getStatusRank(), mapVervPluginStateToPluginStatus(), NodeHeader(), PluginContent(), PluginMatrixProps, sortPluginsByStatus(), Table(), TableProps

### Community 35 - "NetworkTopologyMap.tsx"
Cohesion: 0.29
Nodes (4): SectionLabel(), SectionLabelProps, NodeHealthListProps, NodeProps

### Community 36 - "DeployMenu.tsx"
Cohesion: 0.23
Nodes (9): CreateSmerdRequest, SkeletonDeploymentHistory(), DeployMenu(), DeployMenuProps, TabsOptions, TERMINAL_STATUSES, ListDeploymentsByServiceNameQuery(), DeploymentWidget() (+1 more)

### Community 37 - "scripts"
Cohesion: 0.22
Nodes (9): scripts, build, dev, gen, knip, lint, lint:css, lint:js (+1 more)

### Community 38 - "PluginManageDialog.tsx"
Cohesion: 0.33
Nodes (6): VervPlugin, formatState(), mapStateToStatus(), pluginForms, PluginManageDialog(), UnknownPlugin()

### Community 39 - "DeploymentStatusBadge.tsx"
Cohesion: 0.16
Nodes (13): NodeBaseInfo, NodeStatus, NodesPanel(), NodesPanelProps, plugins, findMasterNodeId(), layoutOuterNodes(), NodeTopologyGraph() (+5 more)

### Community 40 - "compilerOptions"
Cohesion: 0.22
Nodes (8): compilerOptions, allowSyntheticDefaultImports, composite, module, moduleResolution, skipLibCheck, strict, include

### Community 41 - "VCNPeerTable.tsx"
Cohesion: 0.26
Nodes (6): Dialog(), DialogManager, useDialog, Routes, HeadscalePluginForm(), SimplePluginForm()

### Community 42 - "ServiceGraph.tsx"
Cohesion: 0.22
Nodes (9): QueryErrorState(), QueryErrorStateProps, Props, ServiceComingSoon(), Props, Props, ServiceOverviewTab(), ServicePageSkeleton() (+1 more)

### Community 43 - "TasksApi"
Cohesion: 0.29
Nodes (7): InfoMark(), InfoMarkProps, CreateSmerdReq, Smerd, DeployWidget(), DeployWidgetProps, formatStreamStatus()

### Community 45 - "knip.json"
Cohesion: 0.33
Nodes (5): entry, ignore, project, $schema, knip

### Community 46 - "ObservabilityTools.tsx"
Cohesion: 0.40
Nodes (6): Props, TabButton(), ServiceTab, TabInfo, TABS, Props

### Community 47 - "package.json"
Cohesion: 0.40
Nodes (4): name, private, trustedDependencies, type

### Community 50 - "ListDeploymentsResponse"
Cohesion: 0.33
Nodes (8): Port, Volume, Bind, fromProto(), fromProtoPort(), fromProtoPorts(), fromProtoVolume(), fromProtoVolumes()

### Community 59 - "control_plane.ts"
Cohesion: 0.22
Nodes (8): CreateEnvironmentRequest, CreateRegistryRequest, DeleteEnvironmentRequest, DeleteRegistryRequest, EnablePluginRequest, ListPluginsRequest, UpdateEnvironmentRequest, UpdateRegistryRequest

### Community 60 - "Smerds.ts"
Cohesion: 0.29
Nodes (6): createEnvironment, deleteEnvironment, environments, listEnvironments, statefullPlugin, updateEnvironment

### Community 61 - "NodeCard.tsx"
Cohesion: 0.33
Nodes (7): IconButton(), IconButtonProps, isMetricAmber(), isMetricRed(), mapNodeStatus(), NodeCard(), NodeCardProps

### Community 62 - "useCredentialsStore"
Cohesion: 0.40
Nodes (3): FetchNodeHardware(), hardwareQueryOptions(), NodeHardwareQuery()

### Community 63 - "DeploymentFilters.tsx"
Cohesion: 0.47
Nodes (3): BreadcrumbsStore, BreadcrumbsBarProps, Crumb

### Community 64 - ".EnablePlugin"
Cohesion: 0.29
Nodes (3): EnableHeadscaleServer, EnablePluginResponse, EnableStatefullCluster

### Community 65 - "ApiService"
Cohesion: 0.29
Nodes (6): InitReq, Settings, SettingsBase, storeToLocalStorage(), ApiService, withRetries()

### Community 68 - "state.ts"
Cohesion: 0.16
Nodes (9): Button(), ButtonProps, StatefullPgDeadPrompt(), StepsDialogFooterProps, ServicesEmptyState(), ServicesEmptyStateProps, readIncludeInternal(), ServicesPage() (+1 more)

### Community 69 - "CreateAppDialog.tsx"
Cohesion: 0.39
Nodes (5): useSettings(), Props, RegistryImagePicker(), CreateAppDialog(), deriveAppName()

### Community 70 - "ServiceLabelBadge.tsx"
Cohesion: 0.32
Nodes (6): Badge(), BadgeProps, BadgeSpec, ServiceLabelBadge(), ServiceLabelBadgeProps, toBadgeSpec()

### Community 71 - "ObservabilityTools.tsx"
Cohesion: 0.33
Nodes (4): ObservabilityToolsProps, TODO: navigate to tool, TODO: add observability support in Velez, Tool

### Community 72 - "PortsWidget.tsx"
Cohesion: 0.36
Nodes (6): PortMapping(), PortMappingProps, getLinkToPort(), Port, PortsWidget(), PortsWidgetProps

### Community 73 - "VCNPeerTable.tsx"
Cohesion: 0.36
Nodes (5): VCNPeerData, VCNPeerRow(), VCNPeerRowProps, TABLE_HEADERS, VCNPeerTableProps

### Community 75 - "NewServicePage.tsx"
Cohesion: 0.52
Nodes (4): CreateServiceRequest, CreateService(), InitServiceWidget(), InitServiceWidgetProps

### Community 76 - "VolumesWidget.tsx"
Cohesion: 0.38
Nodes (5): ValuesPair(), VolumeMappingProps, Volume, VolumesWidget(), VolumesWidgetProps

### Community 77 - "NetworkTopologyMap.tsx"
Cohesion: 0.29
Nodes (6): NetworkTopologyMap(), NetworkTopologyMapProps, NODE_POSITIONS, STATUS_COLOR, TopologyEdge, TopologyNode

### Community 78 - "DeploymentStatusBadge.tsx"
Cohesion: 0.40
Nodes (5): DeploymentStatus, DeploymentStatusBadge(), DeploymentStatusBadgeProps, STATUS_COLOR, STATUS_DIM

### Community 79 - "ValuesMapping.tsx"
Cohesion: 0.40
Nodes (3): ActionButton(), ActionButtonProps, VolumesWidgetProps

## Knowledge Gaps
- **368 isolated node(s):** `localPlugin`, `$schema`, `entry`, `project`, `ignore` (+363 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **8 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `useToaster` connect `useToaster` to `StatefullPgSettingsScreen.tsx`, `ServiceCard.tsx`, `state.ts`, `CreateAppDialog.tsx`, `DeployMenu.tsx`, `ServicePageModel.ts`, `VCNPeerTable.tsx`, `ServiceGraph.tsx`, `HomePage.tsx`, `TasksApi`, `openStatefullPgDialog.tsx`, `ServiceInfoPage.tsx`, `Sidebar.tsx`, `SmerdPage.tsx`, `tasks.pb.ts`, `TopBar.tsx`?**
  _High betweenness centrality (0.057) - this node is a cross-community bridge._
- **Why does `useDialog` connect `VCNPeerTable.tsx` to `StatefullPgSettingsScreen.tsx`, `PluginMatrix.tsx`, `StepsDialog.tsx`, `CreateAppDialog.tsx`, `state.ts`, `HomePage.tsx`, `openStatefullPgDialog.tsx`, `TopBar.tsx`, `useToaster`?**
  _High betweenness centrality (0.032) - this node is a cross-community bridge._
- **Why does `ServiceApi` connect `ServiceService` to `PluginManageDialog.tsx`, `service_api.pb.ts`, `NewServicePage.tsx`, `api.ts`, `.fetchServiceMetrics`, `.fetchServiceResources`, `service.ts`?**
  _High betweenness centrality (0.027) - this node is a cross-community bridge._
- **What connects `localPlugin`, `$schema`, `entry` to the rest of the system?**
  _373 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ServiceCard.tsx` be split into smaller, more focused modules?**
  _Cohesion score 0.05789235639981909 - nodes in this community are weakly interconnected._
- **Should `fetch.pb.ts` be split into smaller, more focused modules?**
  _Cohesion score 0.04591836734693878 - nodes in this community are weakly interconnected._
- **Should `StepsDialog.tsx` be split into smaller, more focused modules?**
  _Cohesion score 0.07560975609756097 - nodes in this community are weakly interconnected._