// Types corresponding to server/types.go

export interface GetResponse {
  page: Page | null
  folder: FolderEntry[] | null
  allowCreate: boolean
  breadcrumbs: Breadcrumb[]
}

export interface GetAtticListResponse {
  entries: AtticEntry[]
  breadcrumbs: Breadcrumb[]
}

export interface Breadcrumb {
  name: string
  url: string
}

export interface PatchOperation {
  op: 'replace'
  path: string
  value?: any
  from?: string
}

// Types corresponding to storage/model.go

export interface Page {
  url: string
  content: string
  meta: PageMeta
}

export interface PageMeta {
  title: string
  tags: string[] | null
  acls?: AccessRule[] | null
}

export interface FolderEntry {
  url: string
  name: string
  isFolder: boolean
}

export interface AtticEntry {
  rev: number
}

export interface AccessRule {
  subject: string
  ops: AccessOp[] | null
  user?: User
}

export enum AccessOp {
  read = 'read',
  write = 'write',
  delete = 'delete',
}

// export type AccessOp = 'read' | 'write' | 'delete'

export interface User {
  id: string
  username: string
  realName: string
}
