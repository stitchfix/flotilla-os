import axios, { AxiosInstance, AxiosError, AxiosResponse } from "axios"
import * as qs from "qs"
import { has, omit, Omit } from "lodash"
import {
  HTTPMethod,
  CreateTaskPayload,
  RequestArgs,
  Run,
  ListRunParams,
  ListRunResponse,
  RunLog,
  LaunchRequestV2,
  Task,
  ListTaskResponse,
  ListTaskRunsResponse,
  UpdateTaskPayload,
  ListTaskParams,
  ListTaskRunsParams,
  ListClustersResponse,
  ListGroupsResponse,
  ListTagsResponse,
  ListRunEventsResponse,
  RunLogRaw,
  CreateDefinitionTemplatePayload,
} from "../types"

interface IInitOpts {
  baseURL: string
  headers?: object
}

class FlotillaClient {
  private axios: AxiosInstance

  constructor({ baseURL, headers = {} }: IInitOpts) {
    this.axios = axios.create({
      baseURL,
      headers,
      // Note: this is the array format that the Flotilla server accepts.
      paramsSerializer: params =>
        qs.stringify(params, { arrayFormat: "repeat" }),
    })
  }

  /** Requests a task definition. */
  public getTask = ({
    definitionID,
  }: {
    definitionID: string
  }): Promise<Task> =>
    this.request<Task>({
      method: HTTPMethod.GET,
      url: `/v6/task/${definitionID}`,
    })

  /** Requests a task definition by its alias. */
  public getTaskByAlias = ({ alias }: { alias: string }): Promise<Task> =>
    this.request<Task>({
      method: HTTPMethod.GET,
      url: `/v6/task/alias/${alias}`,
    })

  /** Requests a task definition's history. */
  public listTaskRuns = ({
    definitionID,
    params,
  }: {
    definitionID: string
    params: ListTaskRunsParams
  }): Promise<ListTaskRunsResponse> =>
    this.request<ListTaskRunsResponse>({
      method: HTTPMethod.GET,
      url: `/v6/task/${definitionID}/history`,
      params,
    })

  /** Requests a list of task definitions. */
  public listTasks = ({
    params,
  }: {
    params: ListTaskParams
  }): Promise<ListTaskResponse> =>
    this.request<ListTaskResponse>({
      method: HTTPMethod.GET,
      url: `/v6/task`,
      params,
    })

  /** Create a new task definition. */
  public createTask = ({ data }: { data: CreateTaskPayload }): Promise<Task> =>
    this.request<Task>({
      method: HTTPMethod.POST,
      url: `/v6/task`,
      data,
    })

  /** Update an existing task definition. */
  public updateTask = ({
    definitionID,
    data,
  }: {
    definitionID: string
    data: UpdateTaskPayload
  }): Promise<Task> =>
    this.request<Task>({
      method: HTTPMethod.PUT,
      url: `/v6/task/${definitionID}`,
      data,
    })

  /** Delete an existing task definition. */
  public deleteTask = ({
    definitionID,
  }: {
    definitionID: string
  }): Promise<any> =>
    this.request<any>({
      method: HTTPMethod.DELETE,
      url: `/v6/task/${definitionID}`,
    })

  /** Runs a task. */
  public runTask = ({
    definitionID,
    data,
  }: {
    definitionID: string
    data: LaunchRequestV2
  }): Promise<Run> => {
    const d: Omit<LaunchRequestV2, "owner_id"> = omit(data, "owner_id")

    if (has(data, "owner_id")) {
      if (d.run_tags) {
        d.run_tags["OWNER_ID"] = data.owner_id
      } else {
        d.run_tags = { OWNER_ID: data.owner_id }
      }
    }

    return this.request<Run>({
      method: HTTPMethod.PUT,
      url: `/v6/task/${definitionID}/execute`,
      data: d,
    })
  }

  /** Requests list of runs. */
  public listRun = ({
    params,
  }: {
    params: ListRunParams
  }): Promise<ListRunResponse> =>
    this.request<ListRunResponse>({
      method: HTTPMethod.GET,
      url: `/v6/history`,
      params,
    })

  /** Requests a single run. */
  public getRun = ({ runID }: { runID: string }): Promise<Run> =>
    this.request<Run>({
      method: HTTPMethod.GET,
      url: `/v6/task/history/${runID}`,
    })

  /** Requests the logs of a single run. */
  public getRunLog = ({
    runID,
    lastSeen = "",
  }: {
    runID: string
    lastSeen?: string
  }): Promise<RunLog> =>
    this.request<RunLog>({
      method: HTTPMethod.GET,
      url: `/v6/${runID}/logs`,
      params: { last_seen: lastSeen },
    })

  /** Requests the logs of a single run. */
  public getRunLogRaw = ({ runID }: { runID: string }): Promise<RunLogRaw> =>
    this.request<RunLogRaw>({
      method: HTTPMethod.GET,
      url: `/v6/${runID}/logs`,
      params: { raw_text: true },
    })

  /** Stops an existing run */
  public stopRun = ({
    definitionID,
    runID,
  }: {
    definitionID: string
    runID: string
  }): Promise<any> =>
    this.request<any>({
      method: HTTPMethod.DELETE,
      url: `/v6/task/${definitionID}/history/${runID}`,
    })

  /** Requests available clusters. */
  public listClusters = (): Promise<ListClustersResponse> =>
    this.request<ListClustersResponse>({
      method: HTTPMethod.GET,
      url: `/v6/clusters`,
    })

  /** Requests available groups. */
  public listGroups = (): Promise<ListGroupsResponse> =>
    this.request<ListGroupsResponse>({
      method: HTTPMethod.GET,
      url: `/v6/groups`,
      params: { offset: 0, limit: 10000 },
    })

  /** Requests available tags. */
  public listTags = (): Promise<ListTagsResponse> =>
    this.request<ListTagsResponse>({
      method: HTTPMethod.GET,
      url: `/v6/tags`,
      params: { offset: 0, limit: 10000 },
    })

  /** Requests available tags. */
  public listRunEvents = (runID: string): Promise<ListRunEventsResponse> =>
    this.request<ListRunEventsResponse>({
      method: HTTPMethod.GET,
      url: `/v6/${runID}/events`,
    })

  public listDefinitionTemplatesV7 = (): Promise<any> =>
    this.request<any>({
      method: HTTPMethod.GET,
      url: `/v7/template`,
      params: {
        offset: 0,
        limit: 200,
      },
    })

  public getDefinitionTemplateByIdV7 = (id: string): Promise<any> =>
    this.request<any>({
      method: HTTPMethod.GET,
      url: `/v7/template/${id}`,
    })

  public createDefinitionTemplateV7 = (
    data: CreateDefinitionTemplatePayload
  ): Promise<any> =>
    this.request<any>({
      method: HTTPMethod.POST,
      url: `/v7/template`,
      data,
    })

  /** Returns a new Promise that sends an HTTP request when invoked. */
  private request<T>({ method, url, params, data }: RequestArgs): Promise<T> {
    return new Promise((resolve, reject) => {
      this.axios
        .request({ url, method, params, data })
        .then((res: AxiosResponse) => {
          resolve(res.data as T)
        })
        .catch((error: AxiosError) => {
          reject(error)
        })
    })
  }
}

export default FlotillaClient
