export interface Model {
  id: number
  name: string
  provider: string
  apiKey: string
  baseUrl?: string
  modelName: string
  allowedModels?: string
  createdAt?: string
  updatedAt?: string
}

export interface ModelTableData {
  page: number
  pageSize: number
  name?: string
  provider?: string
}
