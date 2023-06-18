// Types corresponding to model/api.go

export interface Breadcrumb {
  name: string
  title: string
  url: string
}

export interface GetAppResponse {
  appTitle: string
  setupMode: boolean
  allowRegister: boolean
  allowAdmin: boolean
  version: string
  gitSha: string
}

export interface PutRequest {
  page?: Page
  folder?: Folder
}

export interface GetContentResponse {
  page: Page | null
  folder: Folder | null
  allowWrite: boolean
  allowDelete: boolean
  breadcrumbs: Breadcrumb[]
}

export interface GetAtticListResponse {
  entries: AtticEntry[]
  breadcrumbs: Breadcrumb[]
}

export interface PatchOperation {
  op: 'replace'
  path: string
  value?: any
  from?: string
}

export interface TokenUserResponse {
  token: string
  user: User
}

// Types corresponding to model/service.go

export interface Page {
  url: string
  content: string
  meta: ContentMeta
}

export interface Folder {
  content: FolderEntry[]
  meta: ContentMeta
}

export interface ContentMeta {
  title: string
  tags: string[] | null
  acl?: AccessRule[] | null
}

export interface FolderEntry {
  url: string
  name: string
  title: string
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
  admin = 'admin',
  register = 'register',
}

export interface User {
  id: string
  username: string
  displayName: string
}

export interface Config {
  appTitle: string
  acl: AccessRule[] | null
}

export interface SearchHit {
  url: string
  meta: ContentMeta
  fragments: Record<string, string[]>
}

// corresponding to service/user_service.go

export const validUsernameRegex = /^[a-zA-Z0-9][a-zA-Z0-9_\\.-]{3,20}$/
