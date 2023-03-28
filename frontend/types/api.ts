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

export interface Page {
  url: string
  meta: { title: string; tags: string[] | null }
  content: string
}

export interface FolderEntry {
  url: string
  name: string
  isFolder: boolean
}

export interface AtticEntry {
  rev: number
}
