export type WorkspaceTabValue = string | string[];

export type WorkspaceTabState = Record<string, WorkspaceTabValue>;

export interface WorkspaceTabRecord {
  id: string;
  routeName: string | null;
  routePath: string;
  routeFullPath: string;
  title: string;
  titleKey?: string;
  titleDefault?: string;
  icon: string | null;
  componentKey: string;
  fixed: boolean;
  closable: boolean;
  cacheable: boolean;
  restorable: boolean;
  query: WorkspaceTabState;
  params: WorkspaceTabState;
  openedAt: number;
  activatedAt: number;
}

export interface WorkspaceTabSnapshot {
  version: 1;
  activeId: string;
  tabs: WorkspaceTabRecord[];
}

export interface WorkspaceTabSyncOptions {
  activate?: boolean;
}
