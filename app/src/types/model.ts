export interface Model {
  id: string
  object: string
  created: number
  owned_by: string
}

export interface GetModelsResponse {
  data: Model[]
}
