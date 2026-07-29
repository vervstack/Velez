# Graph Report - Velez-UI  (2026-07-29)

## Corpus Check
- 153 files · ~93,252 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 940 nodes · 1775 edges · 59 communities (55 shown, 4 thin omitted)
- Extraction: 99% EXTRACTED · 1% INFERRED · 0% AMBIGUOUS · INFERRED: 13 edges (avg confidence: 0.68)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `1f1e1759`
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

## God Nodes (most connected - your core abstractions)
1. `useToaster` - 38 edges
2. `ServiceService` - 24 edges
3. `compilerOptions` - 21 edges
4. `useDialog` - 20 edges
5. `VervPlugin` - 19 edges
6. `ServiceApi` - 17 edges
7. `Button()` - 15 edges
8. `VervPluginType` - 14 edges
9. `VelezAPI` - 13 edges
10. `ControlPlaneService` - 12 edges

## Surprising Connections (you probably didn't know these)
- `openStatefullPgDialog()` --indirect_call--> `StatefullPgSettingsScreen()`  [INFERRED]
  src/dialogs/PluginManageDialog/plugins/openStatefullPgDialog.tsx → src/dialogs/PluginManageDialog/plugins/screens/StatefullPgSettingsScreen.tsx
- `PluginManageDialogProps` --references--> `VervPluginType`  [EXTRACTED]
  src/dialogs/PluginManageDialog/PluginManageDialog.tsx → src/app/api/velez/control_plane_api.pb.ts
- `StepsDialog()` --calls--> `useDialog`  [EXTRACTED]
  src/dialogs/StepsDialog/StepsDialog.tsx → src/app/hooks/dialog/Dialog.tsx
- `ServiceCardActions()` --calls--> `useToaster`  [EXTRACTED]
  src/pages/home/HomePage.tsx → src/app/hooks/toaster/Toaster.ts
- `openStatefullPgDialog()` --indirect_call--> `StatefullPgOverviewScreen()`  [INFERRED]
  src/dialogs/PluginManageDialog/plugins/openStatefullPgDialog.tsx → src/dialogs/PluginManageDialog/plugins/screens/StatefullPgOverviewScreen.tsx

## Import Cycles
- None detected.

## Communities (59 total, 4 thin omitted)

### Community 0 - "ServiceCard.tsx"
Cohesion: 0.07
Nodes (43): Smerd, SmerdStatus, AppCard(), AppCardProps, EnvChip(), EnvChipProps, FreezeChip(), IncidentChip() (+35 more)

### Community 1 - "DeployWidget.tsx"
Cohesion: 0.05
Nodes (42): CreateSmerdRequest, Port, Volume, ActionButton(), ActionButtonProps, Checkbox(), CheckboxProps, InfoMark() (+34 more)

### Community 2 - "Router.tsx"
Cohesion: 0.06
Nodes (36): CreateServiceRequest, ConnectionHealth, ConnectionHealthStatus, useConnectionHealth(), MainLayout(), NAV_TO_ROUTE, NavId, ROUTE_TO_NAV (+28 more)

### Community 3 - "fetch.pb.ts"
Cohesion: 0.05
Nodes (38): b64, b64Encode(), fetchStreamingRequest(), FlattenedRequestPayload, flattenRequestPayload(), getNewLineDelimitedJSONDecodingStream(), getNotifyEntityArrivalSink(), InitReq (+30 more)

### Community 4 - "StepsDialog.tsx"
Cohesion: 0.06
Nodes (21): StatefullPgOverviewScreen(), PipelineFlowLoader(), ScreenSkeletonLoader(), ScreenSkeletonLoaderProps, Step, StepScreenProps, StepsDialog(), StepsDialogProps (+13 more)

### Community 5 - "ServicePageModel.ts"
Cohesion: 0.09
Nodes (24): EnvCard(), EnvCardProps, ServiceEnvironment, ServiceGraphNode, ServiceResource, VervonomiconDocs, ListEnvironmentsQuery(), useGetServiceResourcesQuery() (+16 more)

### Community 6 - "service_api.pb.ts"
Cohesion: 0.06
Nodes (34): AboutService, Absent, BaseCreateDeployRequest, BaseGetServiceResponse, BoundResource, CreateDeploy, CreateDeployRequestUpgrade, CreateDeployResponse (+26 more)

### Community 7 - "ServiceService"
Cohesion: 0.10
Nodes (7): ListServicesRequest, ListServicesResponse, ServiceApi, VervAppService, ServiceAbout, ServiceGraphData, ServiceService

### Community 8 - "control_plane_api.pb.ts"
Cohesion: 0.07
Nodes (26): Absent, BaseEnableHeadscaleServer, BaseEnablePluginRequest, ConnectSlave, ConnectSlaveRequest, ConnectSlaveResponse, EnableHeadscaleServerDeployHeadscaleConfig, EnableHeadscaleServerExternalHeadscaleConnection (+18 more)

### Community 9 - "velez_api.pb.ts"
Cohesion: 0.07
Nodes (26): AssembleConfig, AssembleConfigRequest, AssembleConfigResponse, BreakConnections, BreakConnectionsRequest, BreakConnectionsResponse, CreateSmerd, DropSmerd (+18 more)

### Community 10 - "compilerOptions"
Cohesion: 0.08
Nodes (24): compilerOptions, allowImportingTsExtensions, allowJs, allowSyntheticDefaultImports, esModuleInterop, forceConsistentCasingInFileNames, isolatedModules, jsx (+16 more)

### Community 11 - "HomePage.tsx"
Cohesion: 0.11
Nodes (13): ServiceBaseInfo, SkeletonLoader(), SkeletonLoaderProps, SkeletonServiceCard(), SkeletonSmerdRow(), formatLastDeployed(), HomePage(), ServiceCardActions() (+5 more)

### Community 12 - "devDependencies"
Cohesion: 0.09
Nodes (23): devDependencies, eslint, eslint-import-resolver-typescript, @eslint/js, eslint-plugin-import-x, eslint-plugin-react, eslint-plugin-react-hooks, eslint-plugin-react-refresh (+15 more)

### Community 13 - "index.ts"
Cohesion: 0.23
Nodes (11): EnableHeadscaleServer, VervPluginState, VervPluginType, Routes, Button(), ButtonProps, StatefullPgDeadPromptProps, metaByType (+3 more)

### Community 14 - "openStatefullPgDialog.tsx"
Cohesion: 0.15
Nodes (11): Dialog(), DialogManager, useDialog, queryClient, HeadscalePluginForm(), openStatefullPgDialog(), StatefullPgReviewScreen(), SimplePluginForm() (+3 more)

### Community 15 - "ServiceInfoPage.tsx"
Cohesion: 0.14
Nodes (15): TagChip(), TagChipProps, ActionsRowProps, ServiceInfoPage(), ServicePageHeader(), ServiceTab, ServiceTagsStrip(), ServiceTagsStripProps (+7 more)

### Community 16 - "service.ts"
Cohesion: 0.13
Nodes (17): CreateDeployRequest, GetServiceEnvironmentsRequest, GetServiceGraphRequest, GetServiceMetricsRequest, GetServiceRequest, ListDeploymentsRequest, RemoveServiceRequest, RestartServiceRequest (+9 more)

### Community 17 - "velez_common.pb.ts"
Cohesion: 0.12
Nodes (16): ConfigFormat, Connection, Container, ContainerHardware, ContainerHealthcheck, ContainerSettings, FileConfig, Image (+8 more)

### Community 18 - "Sidebar.tsx"
Cohesion: 0.14
Nodes (11): SkeletonNodeRow(), ListNodesQuery(), NAV_ITEMS, NavId, NavItemProps, NodesList(), Sidebar(), SidebarProps (+3 more)

### Community 19 - "SmerdPage.tsx"
Cohesion: 0.15
Nodes (7): BreadcrumbsBarProps, Crumb, QueryErrorState(), QueryErrorStateProps, formatTimestamp(), SmerdMetaSection(), SmerdPage()

### Community 20 - "Architecture"
Cohesion: 0.14
Nodes (12): API calls, Architecture, Coding Rules, Commands, Environment, Exploration Rules, Layer structure, Proto regeneration (+4 more)

### Community 21 - "ControlPlaneService"
Cohesion: 0.19
Nodes (4): ControlPlaneAPI, EnableStatefullCluster, ListEnvironmentsResponse, ControlPlaneService

### Community 22 - "dependencies"
Cohesion: 0.15
Nodes (13): dependencies, classnames, framer-motion, @microlink/react-json-view, react, react-dom, react-router-dom, react-tooltip (+5 more)

### Community 23 - "ControlPlanePage.tsx"
Cohesion: 0.22
Nodes (8): ListNodesResponse, SkeletonNodeCard(), ControlPlanePage(), NodeListProps, PluginsListProps, StatsGridProps, NodeHardwareQuery(), PluginMatrix()

### Community 24 - "tasks.pb.ts"
Cohesion: 0.15
Nodes (12): AssembleConfigTaskPayload, ConnectServiceToVpnTaskPayload, CopyToVolumeTaskPayload, CreateServiceTaskPayload, CreateSmerdTaskPayload, DropSmerdTaskPayload, EnableStatefullTaskPayload, TaskStatus (+4 more)

### Community 25 - "velez.ts"
Cohesion: 0.26
Nodes (10): GetHardwareResponse, ListSmerdsRequest, ListSmerdsResponse, SearchImagesResponse, FetchSmerd(), FetchSmerds(), FetchSmerdsByServiceId(), GetSmerd() (+2 more)

### Community 26 - "VelezAPI"
Cohesion: 0.15
Nodes (3): VelezAPI, FetchNodeHardware(), ListImages()

### Community 27 - "NodeCard.tsx"
Cohesion: 0.23
Nodes (9): NodeBaseInfo, NodeStatus, isMetricAmber(), isMetricRed(), mapNodeStatus(), NodeCard(), NodeCardProps, NodeHealthListProps (+1 more)

### Community 28 - "VervClosedNetworkPage.tsx"
Cohesion: 0.21
Nodes (10): Level, StatCard(), StatCardProps, CodeBlock(), CodeBlockProps, tokenizeLine(), MOCK_EDGES, MOCK_NODES (+2 more)

### Community 29 - "TopBar.tsx"
Cohesion: 0.27
Nodes (7): IsStatefullModeEnabled(), ListPluginsQuery(), LeftSideProps, NavId, RightZone(), SingleNodeStub(), TopBarProps

### Community 30 - "DeployRow.tsx"
Cohesion: 0.31
Nodes (8): DeploymentInfo, DeploymentHistoryProps, DeployRow(), DeployRowProps, formatTimestamp(), getStatusClass(), parseImageTag(), TODO: replace with git commit hash when GetServiceGraph adds git field

### Community 31 - "useToaster"
Cohesion: 0.31
Nodes (8): Toast, Toaster, useToaster, StatefullPgDeadPrompt(), RemoveServiceDialog(), RemoveServiceDialogProps, Toast(), Toaster()

### Community 32 - "StatefullPgSettingsScreen.tsx"
Cohesion: 0.24
Nodes (5): ChoiceProps, StatefullPgReviewScreenProps, StatefullPgSettingsScreen(), StatefullPgSettingsScreenProps, StatefullPgContext

### Community 33 - "services.ts"
Cohesion: 0.29
Nodes (9): LIST_REQ, useGetServiceAboutQuery(), useGetServiceMetricsQuery(), MetricTile(), MetricTileProps, progressBarFillClass(), ServiceHero(), ServiceHeroProps (+1 more)

### Community 34 - "PluginMatrix.tsx"
Cohesion: 0.31
Nodes (8): IconButton(), IconButtonProps, mapVervPluginStateToPluginStatus(), NodeHeader(), PluginContent(), PluginMatrixProps, Table(), TableProps

### Community 35 - "NetworkTopologyMap.tsx"
Cohesion: 0.22
Nodes (8): SectionLabel(), SectionLabelProps, NetworkTopologyMap(), NetworkTopologyMapProps, NODE_POSITIONS, STATUS_COLOR, TopologyEdge, TopologyNode

### Community 36 - "DeployMenu.tsx"
Cohesion: 0.29
Nodes (7): SkeletonDeploymentHistory(), DeployMenu(), DeployMenuProps, TabsOptions, TERMINAL_STATUSES, ActionsRow(), ListDeploymentsByServiceNameQuery()

### Community 37 - "scripts"
Cohesion: 0.22
Nodes (9): scripts, build, dev, gen, knip, lint, lint:css, lint:js (+1 more)

### Community 38 - "PluginManageDialog.tsx"
Cohesion: 0.31
Nodes (7): VervPlugin, formatState(), mapStateToStatus(), pluginForms, PluginManageDialog(), PluginManageDialogProps, UnknownPlugin()

### Community 39 - "DeploymentStatusBadge.tsx"
Cohesion: 0.28
Nodes (7): DeploymentStatus, Badge(), BadgeProps, DeploymentStatusBadge(), DeploymentStatusBadgeProps, STATUS_COLOR, STATUS_DIM

### Community 40 - "compilerOptions"
Cohesion: 0.22
Nodes (8): compilerOptions, allowSyntheticDefaultImports, composite, module, moduleResolution, skipLibCheck, strict, include

### Community 41 - "VCNPeerTable.tsx"
Cohesion: 0.36
Nodes (5): VCNPeerData, VCNPeerRow(), VCNPeerRowProps, TABLE_HEADERS, VCNPeerTableProps

### Community 42 - "ServiceGraph.tsx"
Cohesion: 0.46
Nodes (7): useGetServiceGraphQuery(), labelLines(), nodeEdgeX(), renderNode(), ServiceGraph(), ServiceGraphProps, spreadY()

### Community 43 - "TasksApi"
Cohesion: 0.29
Nodes (4): TasksApi, toProto(), DeploySmerd(), DeploySmerdStream()

### Community 44 - "api.ts"
Cohesion: 0.48
Nodes (6): GetInitReq(), InitReq, keyPath(), pathPrefixPath(), StoreApiKey(), StorePathPrefix()

### Community 45 - "knip.json"
Cohesion: 0.33
Nodes (5): entry, ignore, project, $schema, knip

### Community 46 - "ObservabilityTools.tsx"
Cohesion: 0.33
Nodes (4): ObservabilityToolsProps, TODO: navigate to tool, TODO: add observability support in Velez, Tool

### Community 47 - "package.json"
Cohesion: 0.40
Nodes (4): name, private, trustedDependencies, type

## Knowledge Gaps
- **338 isolated node(s):** `localPlugin`, `$schema`, `entry`, `project`, `ignore` (+333 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **4 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `useToaster` connect `useToaster` to `ServiceCard.tsx`, `services.ts`, `DeployWidget.tsx`, `Router.tsx`, `DeployMenu.tsx`, `ServiceGraph.tsx`, `HomePage.tsx`, `index.ts`, `openStatefullPgDialog.tsx`, `ServiceInfoPage.tsx`, `SmerdPage.tsx`, `velez.ts`?**
  _High betweenness centrality (0.041) - this node is a cross-community bridge._
- **Why does `useDialog` connect `openStatefullPgDialog.tsx` to `PluginMatrix.tsx`, `StepsDialog.tsx`, `DeployMenu.tsx`, `HomePage.tsx`, `index.ts`, `ServiceInfoPage.tsx`?**
  _High betweenness centrality (0.032) - this node is a cross-community bridge._
- **Why does `ServiceService` connect `ServiceService` to `services.ts`, `Router.tsx`, `DeployMenu.tsx`, `ServicePageModel.ts`, `HomePage.tsx`, `index.ts`, `ServiceInfoPage.tsx`, `service.ts`, `.fetchServiceMetrics`, `ListDeploymentsResponse`, `useToaster`?**
  _High betweenness centrality (0.024) - this node is a cross-community bridge._
- **What connects `localPlugin`, `$schema`, `entry` to the rest of the system?**
  _343 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ServiceCard.tsx` be split into smaller, more focused modules?**
  _Cohesion score 0.06502816180235535 - nodes in this community are weakly interconnected._
- **Should `DeployWidget.tsx` be split into smaller, more focused modules?**
  _Cohesion score 0.05129561078794289 - nodes in this community are weakly interconnected._
- **Should `Router.tsx` be split into smaller, more focused modules?**
  _Cohesion score 0.05909090909090909 - nodes in this community are weakly interconnected._