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

const mock = new MockAdapter(axios)
const MOCK_TASK: Task = {
  env: [{ name: "foo", value: "bar" }],
  arn: "my_arn",
  definition_id: "my_definition_id",
  image: "my_image",
  group_name: "my_group_name",
  container_name: "my_container_name",
  alias: "my_alias",
  memory: 512,
  command: "my_command",
  tags: ["tag_one", "tag_two"],
}

const MOCK_RUN: Run = {
  instance: {
    dns_name: "my_dns_name",
    instance_id: "my_instance_id",
  },
  task_arn: "my_task_arn",
  run_id: "my_run_id",
  definition_id: "my_definition_id",
  alias: "my_alias",
  image: "my_image",
  cluster: "my_cluster",
  exit_code: 1,
  status: RunStatus.STOPPED,
  started_at: "2019-05-02T19:26:21.559Z",
  finished_at: "2019-05-02T20:21:48.36Z",
  group_name: "my_group_name",
  env: [{ name: "foo", value: "bar" }],
}

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
      definitions: [MOCK_TASK],
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
    mock.onGet(`/v1/task/${id}`).reply(200, MOCK_TASK)
    expect(await client.getTask({ definitionID: id })).toEqual(MOCK_TASK)
  })

  it("getTaskByAlias", async () => {
    const alias = "my_task_alias"
    mock.onGet(`/v1/task/alias/${alias}`).reply(200, MOCK_TASK)
    expect(await client.getTaskByAlias({ alias })).toEqual(MOCK_TASK)
  })

  it("getTaskHistory", async () => {
    const id = "my_task"
    const res: ListTaskRunsResponse = {
      history: [MOCK_RUN],
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
      command: "command",
      tags: ["tag_one"],
    }
    const res: Task = {
      ...data,
      arn: "arn",
      definition_id: "definition_id",
      container_name: "container_name",
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
      command: "command",
      tags: ["tag_one"],
    }
    const res: Task = {
      ...data,
      alias: "alias",
      arn: "arn",
      definition_id: "definition_id",
      container_name: "container_name",
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

    mock.onPut(`/v1/task/${id}/execute`).reply(200, MOCK_RUN)
    expect(await client.runTask({ definitionID: id, data })).toEqual(MOCK_RUN)
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
      history: [MOCK_RUN],
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
    mock.onGet(`/v1/task/history/${runID}`).reply(200, MOCK_RUN)
    expect(await client.getRun({ runID })).toEqual(MOCK_RUN)
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
