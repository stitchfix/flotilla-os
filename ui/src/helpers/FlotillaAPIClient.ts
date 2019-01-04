import axios, { AxiosRequestConfig, AxiosResponse, AxiosError } from "axios"
import { get, isEmpty } from "lodash"
import * as qs from "qs"
import * as urljoin from "url-join"
import {
  IFlotillaAPIError,
  IFlotillaCreateTaskPayload,
  IFlotillaRunTaskPayload,
  IFlotillaEditTaskPayload,
  IFlotillaEnv,
  IFlotillaRun,
  IFlotillaTaskDefinition,
} from "../../index"
import { stringToSelectOpt } from "./reactSelectHelpers"

export interface IRequestOpts {
  method: string
  path: string
  payload?: object
  preprocess?: (pre: any) => any
  query?: object
}

export interface IConstructURLOpts {
  path: string
  query?: object
}

class FlotillaAPIClient {
  location: string

  constructor(location: string) {
    this.location = location
  }

  /** Requests list of task definitions. */
  getTasks = (opts: { query: object }): Promise<any> => {
    return this.request({
      method: "get",
      path: "/v1/task",
      query: opts.query,
    })
  }

  /** Requests data for a specific task definition. */
  getTask = ({ definitionID }: { definitionID: string }): Promise<any> => {
    return this.request({
      method: "get",
      path: `/v1/task/${definitionID}`,
      preprocess: FlotillaAPIClient.preprocessTaskDefinitionResponse,
    })
  }

  /** Requests data for a specific task definition by its alias. */
  getTaskByAlias = ({ alias }: { alias: string }): Promise<any> => {
    return this.request({
      method: "get",
      path: `/v1/task/alias/${alias}`,
      preprocess: FlotillaAPIClient.preprocessTaskDefinitionResponse,
    })
  }

  /** Requests the history of a task. */
  getTaskHistory = ({
    definitionID,
    query,
  }: {
    definitionID: string
    query: object | undefined
  }): Promise<any> => {
    return this.request({
      method: "get",
      path: `/v1/task/${definitionID}/history`,
      query,
    })
  }

  /** Creates task definition */
  createTask = ({
    values,
  }: {
    values: IFlotillaCreateTaskPayload
  }): Promise<any> => {
    return this.request({
      method: "post",
      path: "/v1/task",
      payload: values,
    })
  }

  /** Updates a task definition. */
  updateTask = ({
    definitionID,
    values,
  }: {
    definitionID: string
    values: IFlotillaEditTaskPayload
  }): Promise<any> => {
    return this.request({
      method: "put",
      path: `/v1/task/${definitionID}`,
      payload: values,
    })
  }

  /** Deletes a task definition. */
  deleteTask = ({ definitionID }: { definitionID: string }): Promise<any> => {
    return this.request({
      method: "delete",
      path: `/v1/task/${definitionID}`,
    })
  }

  /** Executes a task definition. */
  runTask = ({
    definitionID,
    values,
  }: {
    definitionID: string
    values: IFlotillaRunTaskPayload
  }): Promise<any> => {
    let _values: any = values

    if (values.run_tags) {
      _values = {
        ...values,
        run_tags: FlotillaAPIClient.transformUIRunTagsToAPIRunTags(
          values.run_tags
        ),
      }
    }

    return this.request({
      method: "put",
      path: `/v4/task/${definitionID}/execute`,
      payload: _values,
    })
  }

  /** Requests list of currently running tasks. */
  getActiveRuns = (opts: { query: object }): Promise<any> => {
    return this.request({
      method: "get",
      path: "/v1/history",
      query: opts.query,
    })
  }

  /** Terminates a currently running task. */
  stopRun = ({
    definitionID,
    runID,
  }: {
    definitionID: string
    runID: string
  }): Promise<any> => {
    return this.request({
      method: "delete",
      path: `/v1/task/${definitionID}/history/${runID}`,
    })
  }

  /** Gets data for a specific run. */
  getRun = ({ runID }: { runID: string }): Promise<any> => {
    return this.request({
      method: "get",
      path: `/v1/task/history/${runID}`,
    })
  }

  /** Gets logs of a specific run. */
  getRunLogs = ({ runID, lastSeen }: { runID: string; lastSeen?: string }) => {
    return this.request({
      method: "get",
      path: `/v1/${runID}/logs`,
      query: { last_seen: lastSeen },
    })
  }

  /** Requests the available Flotilla groups. */
  getGroups = (): Promise<any> => {
    return this.request({
      method: "get",
      path: `/v1/groups`,
      query: { limit: 2000 },
      preprocess: (res: any): string[] =>
        get(res, "groups", [])
          .filter((v: string) => !isEmpty(v))
          .map(stringToSelectOpt),
    })
  }

  /** Requests the available ECS clusters. */
  getClusters = () => {
    return this.request({
      method: "get",
      path: `/v1/clusters`,
      preprocess: (res: any): string[] =>
        get(res, "clusters", [])
          .filter((v: string) => !isEmpty(v))
          .map(stringToSelectOpt),
    })
  }

  /** Requests lists of existing tags. */
  getTags = () => {
    return this.request({
      method: "get",
      path: "/v1/tags",
      query: { limit: 5000 },
      preprocess: (res: any): string[] =>
        get(res, "tags", [])
          .filter((v: string) => !isEmpty(v))
          .map(stringToSelectOpt),
    })
  }

  request = (opts: IRequestOpts): Promise<any> => {
    const { method, path, payload, preprocess, query } = opts
    return new Promise((resolve, reject) => {
      let config: AxiosRequestConfig = {
        method,
        url: this.constructURL({ path, query }),
      }

      if (!!payload && !isEmpty(payload)) {
        config.data = payload
      }

      axios(config)
        .then((res: AxiosResponse) => {
          if (!!preprocess) {
            resolve(preprocess(res.data))
          }

          resolve(res.data)
        })
        .catch((e: AxiosError) => {
          const error: IFlotillaAPIError = this.processError(e)
          reject(error)
        })
    })
  }

  constructURL = (opts: IConstructURLOpts): string => {
    let ret = `${urljoin(this.location, opts.path)}`

    if (!!opts.query) {
      ret += `?${qs.stringify(opts.query, { indices: false })}`
    }

    return ret
  }

  processError = (error: AxiosError): IFlotillaAPIError => {
    if (!!error.response) {
      return {
        data: get(error, ["response", "data"]),
        status: error.response.status,
        headers: error.response.headers,
      }
    }

    return { data: error.message }
  }

  static transformUIRunTagsToAPIRunTags = (
    arr: IFlotillaEnv[]
  ): { [key: string]: any } => {
    return arr.reduce((acc: any, val) => {
      acc[val.name] = val.value
      return acc
    }, {})
  }

  static preprocessTaskDefinitionResponse = (
    data: IFlotillaTaskDefinition
  ) => ({
    ...data,
    tags: get(data, "tags", []).filter((t: string) => t.length > 0),
  })
}

export default FlotillaAPIClient
