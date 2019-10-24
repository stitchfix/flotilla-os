import axios from "axios"
import MockAdapter from "axios-mock-adapter"
import FlotillaClient from "../FlotillaClient"
import {
  Task,
  Run,
  RunStatus,
  ListTaskResponse,
  SortOrder,
  CreateTaskPayload,
  UpdateTaskPayload,
  ListTaskRunsResponse,
  RunTaskPayload,
  ListRunParams,
  ListRunResponse,
  RunLog,
} from "../../types"
import { createMockRunObject, createMockTaskObject } from "../testHelpers"

const mock = new MockAdapter(axios)

describe("FlotillaClient", () => {
  let client: FlotillaClient

  beforeAll(() => {
    client = new FlotillaClient({ baseURL: "" })
  })

  afterAll(() => {
    mock.reset()
  })

  afterEach(() => {
    mock.restore()
  })

  // ---------------------------------------------------------------------------
  // Task-related endpoints
  // ---------------------------------------------------------------------------
  it("getTasks", async () => {
    const res: ListTaskResponse = {
      definitions: [createMockTaskObject()],
      total: 1,
      offset: 0,
      limit: 20,
      sort_by: "alias",
      order: SortOrder.ASC,
    }
    mock.onGet(`/v1/task`).reply(200, res)
    expect(
      await client.listTasks({ params: { offset: 0, limit: 20 } })
    ).toEqual(res)
  })

  it("getTask", async () => {
    const id = "my_task"
    mock.onGet(`/v1/task/${id}`).reply(200, createMockTaskObject())
    expect(await client.getTask({ definitionID: id })).toEqual(
      createMockTaskObject()
    )
  })

  it("getTaskByAlias", async () => {
    const alias = "my_task_alias"
    mock.onGet(`/v1/task/alias/${alias}`).reply(200, createMockTaskObject())
    expect(await client.getTaskByAlias({ alias })).toEqual(
      createMockTaskObject()
    )
  })

  it("getTaskHistory", async () => {
    const id = "my_task"
    const res: ListTaskRunsResponse = {
      history: [createMockRunObject()],
      total: 1,
      offset: 0,
      limit: 20,
      sort_by: "alias",
      order: SortOrder.ASC,
    }
    mock.onGet(`/v1/task/${id}/history`).reply(200, res)
    expect(
      await client.listTaskRuns({
        definitionID: id,
        params: { offset: 0, limit: 20 },
      })
    ).toEqual(res)
  })

  it("createTask", async () => {
    const data: CreateTaskPayload = {
      env: [],
      image: "image",
      group_name: "group_name",
      alias: "alias",
      memory: 1000,
      cpu: 1000,
      command: "command",
      tags: ["tag_one"],
    }
    const res: Task = {
      ...data,
      arn: "arn",
      definition_id: "definition_id",
      container_name: "container_name",
      privileged: false,
    }
    mock.onPost(`/v1/task`).reply(200, res)
    expect(await client.createTask({ data })).toEqual(res)
  })

  it("updateTask", async () => {
    const id = "my_task"
    const data: UpdateTaskPayload = {
      env: [],
      image: "image",
      group_name: "group_name",
      memory: 1000,
      cpu: 1000,
      command: "command",
      tags: ["tag_one"],
    }
    const res: Task = {
      ...data,
      alias: "alias",
      arn: "arn",
      definition_id: "definition_id",
      container_name: "container_name",
      privileged: false,
    }
    mock.onPut(`/v1/task/${id}`).reply(200, res)
    expect(await client.updateTask({ definitionID: id, data })).toEqual(res)
  })

  it("deleteTask", async () => {
    const id = "my_task"
    const res = {}
    mock.onDelete(`/v1/task/${id}`).reply(200, res)
    expect(await client.deleteTask({ definitionID: id })).toEqual(res)
  })

  it("runTask", async () => {
    const id = "my_task"
    const data: RunTaskPayload = {
      cluster: "cluster",
      env: [],
      run_tags: {},
    }

    mock.onPut(`/v4/task/${id}/execute`).reply(200, createMockRunObject())
    expect(await client.runTask({ definitionID: id, data })).toEqual(
      createMockRunObject()
    )
  })

  // ---------------------------------------------------------------------------
  // Run-related endpoints
  // ---------------------------------------------------------------------------
  it("listRun", async () => {
    const params: ListRunParams = {
      offset: 0,
      limit: 20,
    }
    const res: ListRunResponse = {
      history: [createMockRunObject()],
      offset: 0,
      limit: 20,
      sort_by: "started_at",
      order: SortOrder.ASC,
      total: 1,
    }

    mock.onGet(`/v1/history`).reply(200, res)
    expect(await client.listRun({ params })).toEqual(res)
  })

  it("getRun", async () => {
    const runID = "run_id"
    mock.onGet(`/v1/task/history/${runID}`).reply(200, createMockRunObject())
    expect(await client.getRun({ runID })).toEqual(createMockRunObject())
  })

  it("getRunLogs", async () => {
    const runID = "run_id"
    const lastSeen = ""
    const res: RunLog = {
      log: "log",
      last_seen: "last_seen",
    }
    mock.onGet(`/v1/${runID}/logs`).reply(200, res)
    expect(await client.getRunLog({ runID, lastSeen })).toEqual(res)
  })

  it("stopRun", async () => {
    const definitionID = "definition_id"
    const runID = "run_id"

    mock.onDelete(`/v1/task/${definitionID}/history/${runID}`).reply(200, {})
    expect(await client.stopRun({ runID, definitionID })).toEqual({})
  })

  // ---------------------------------------------------------------------------
  // Misc endpoints
  // ---------------------------------------------------------------------------
  it("getClusters", async () => {
    const res = { clusters: [] }
    mock.onGet(`/v1/clusters`).reply(200, res)
    expect(await client.listClusters()).toEqual(res)
  })

  it("getTags", async () => {
    const res = { tags: [], offset: 0, limit: 20, total: 0 }
    mock.onGet(`/v1/tags`).reply(200, res)
    expect(await client.listTags()).toEqual(res)
  })

  it("getGroups", async () => {
    const res = { groups: [], offset: 0, limit: 20, total: 0 }
    mock.onGet(`/v1/groups`).reply(200, res)
    expect(await client.listGroups()).toEqual(res)
  })
})
