export interface IFlotillaAPIError {
  data: any
  status?: any
  headers?: any
}

/** The values required to create a task definition. */
export interface IFlotillaCreateTaskPayload {}

/** The values required to execute a task definition. */
export interface IFlotillaRunTaskPayload {}
