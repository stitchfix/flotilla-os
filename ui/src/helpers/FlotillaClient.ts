import axios, { AxiosInstance, AxiosError, AxiosResponse } from "axios"
import * as qs from "qs"
import cookie from "cookie"
import { get } from "lodash"
import {
  HTTPMethod,
  CreateTaskPayload,
  RequestArgs,
  Run,
  ListRunParams,
  ListRunResponse,
  RunLog,
  RunTaskPayload,
  Task,
  ListTaskResponse,
  ListTaskRunsResponse,
  UpdateTaskPayload,
  ListTaskParams,
  ListTaskRunsParams,
  ListClustersResponse,
  ListGroupsResponse,
  ListTagsResponse,
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
      url: `/v1/task/${definitionID}`,
    })

  /** Requests a task definition by its alias. */
  public getTaskByAlias = ({ alias }: { alias: string }): Promise<Task> =>
    this.request<Task>({
      method: HTTPMethod.GET,
      url: `/v1/task/alias/${alias}`,
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
      url: `/v1/task/${definitionID}/history`,
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
      url: `/v1/task`,
      params,
    })

  /** Create a new task definition. */
  public createTask = ({ data }: { data: CreateTaskPayload }): Promise<Task> =>
    this.request<Task>({
      method: HTTPMethod.POST,
      url: `/v1/task`,
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
      url: `/v1/task/${definitionID}`,
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
      url: `/v1/task/${definitionID}`,
    })

  /** Runs a task. */
  public runTask = ({
    definitionID,
    data,
  }: {
    definitionID: string
    data: RunTaskPayload
  }): Promise<Run> => {
    let d: RunTaskPayload = data

    // Get owner ID.
    let ownerID: string = "flotilla-ui"

    if (process.env.REACT_APP_RUN_TAG_OWNER_ID_COOKIE_PATH) {
      console.log("did set cookie env")
      console.log(process.env.REACT_APP_RUN_TAG_OWNER_ID_COOKIE_PATH)
      const cookies = cookie.parse(document.cookie)
      console.log(cookies)
      ownerID = get(
        cookies,
        process.env.REACT_APP_RUN_TAG_OWNER_ID_COOKIE_PATH,
        "flotilla-ui"
      )
      console.log(`ownerID: ${ownerID}`)
    }

    d.run_tags = {
      ...d.run_tags,
      OWNER_ID: ownerID,
    }

    console.log(d.run_tags)

    return this.request<Run>({
      method: HTTPMethod.PUT,
      url: `/v4/task/${definitionID}/execute`,
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
      url: `/v1/history`,
      params,
    })

  /** Requests a single run. */
  public getRun = ({ runID }: { runID: string }): Promise<Run> =>
    this.request<Run>({
      method: HTTPMethod.GET,
      url: `/v1/task/history/${runID}`,
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
      url: `/v1/${runID}/logs`,
      params: { last_seen: lastSeen },
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
      url: `/v1/task/${definitionID}/history/${runID}`,
    })

  /** Requests available clusters. */
  public listClusters = (): Promise<ListClustersResponse> =>
    this.request<ListClustersResponse>({
      method: HTTPMethod.GET,
      url: `/v1/clusters`,
    })

  /** Requests available groups. */
  public listGroups = (): Promise<ListGroupsResponse> =>
    this.request<ListGroupsResponse>({
      method: HTTPMethod.GET,
      url: `/v1/groups`,
      params: { offset: 0, limit: 10000 },
    })

  /** Requests available tags. */
  public listTags = (): Promise<ListTagsResponse> =>
    this.request<ListTagsResponse>({
      method: HTTPMethod.GET,
      url: `/v1/tags`,
      params: { offset: 0, limit: 10000 },
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
