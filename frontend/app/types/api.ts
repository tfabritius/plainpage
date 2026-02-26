/* Types corresponding to backend/model/api.go */

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
  version?: string
  gitSha?: string
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
  value?: unknown
  from?: string
}

export interface LoginRequest {
  username: string
  password: string
}

export interface ChangePasswordRequest {
  currentPassword: string
  newPassword: string
}

export interface PostUserRequest {
  username: string
  password: string
  displayName: string
}

export interface DeleteUserRequest {
  password: string
}

export interface LoginResponse {
  accessToken: string
  user: User
}

export interface RefreshResponse {
  accessToken: string
  user: User
}

export interface AtticEntry {
  rev: number
}

export interface TrashEntry {
  url: string
  deletedAt: number
  meta: ContentMeta
}

export interface GetTrashListResponse {
  items: TrashEntry[]
  totalCount: number
  page: number
  limit: number
}

export interface GetTrashPageResponse {
  page: Page
}

export interface TrashActionRequest {
  items: TrashItemRef[]
}

export interface TrashItemRef {
  url: string
  deletedAt: number
}

export interface Page {
  url: string
  content: string
  meta: ContentMeta
}

export interface Folder {
  url: string
  content: FolderEntry[]
  meta: ContentMeta
}

export interface ContentMeta {
  title: string
  tags: string[] | null
  acl?: AccessRule[] | null
  modifiedAt?: string
  modifiedByUsername?: string
  modifiedByDisplayName?: string
}

export interface FolderEntry {
  url: string
  name: string
  title: string
  isFolder: boolean
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

// Subsets for validation purposes
export const ContentAccessOps = [AccessOp.read, AccessOp.write, AccessOp.delete] as const
export const ConfigAccessOps = [AccessOp.admin, AccessOp.register] as const

// Type helpers for compile-time type checking
export type ContentAccessOp = (typeof ContentAccessOps)[number]
export type ConfigAccessOp = (typeof ConfigAccessOps)[number]

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
  isFolder: boolean
}

export interface SearchResponse {
  items: SearchHit[]
  page: number
  limit: number
  hasMore: boolean
}

// corresponding to service/user_service.go

export const validUsernameRegex = /^[a-z0-9][a-z0-9_\\.-]{3,20}$/i

// correspongin to server/content.go

export const validUrlPartRegex = /^[a-z0-9-][a-z0-9_-]*$/
