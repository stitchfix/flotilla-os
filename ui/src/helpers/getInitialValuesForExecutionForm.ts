import { get } from "lodash"
import getOwnerIdRunTagFromCookies from "./getOwnerIdRunTagFromCookies"
import {
  Executable,
  LaunchRequestV2,
  Run,
  Task,
  Template,
  TemplateExecutionRequest,
  ExecutionRequestCommon,
  ExecutionEngine,
  Env,
  DefaultNodeLifecycle,
  DefaultExecutionEngine,
} from "../types"
import constructDefaultObjectFromJsonSchema from "./constructDefaultObjectFromJsonSchema"

export function getInitialValuesForTaskExecutionForm(
  t: Task,
  r: Run | null
): LaunchRequestV2 {
  const common = getInitialValuesForCommonExecutionFields(t, r)

  // Set command value.
  const command: string = r && r.command ? r.command : t.command

  common.command = command

  return common
}

export function getInitialValuesForTemplateExecutionForm(
  t: Template,
  r: Run | null
): TemplateExecutionRequest {
  const req: TemplateExecutionRequest = {
    ...getInitialValuesForCommonExecutionFields(t, r),
    template_payload: get(
      r,
      ["execution_request_custom", "template_payload"],
      constructDefaultObjectFromJsonSchema(t.schema)
    ),
  }

  return req
}

function getInitialValuesForCommonExecutionFields(
  e: Executable,
  r: Run | null
): ExecutionRequestCommon {
  // Set ownerID value.
  const ownerID = get(
    r,
    ["run_tags", "owner_id"],
    getOwnerIdRunTagFromCookies()
  )

  // Set env value.
  let env: Env[] | null = r && r.env ? r.env : e.env

  // Filter out invalid run env if specified in dotenv file.
  if (env === null) {
    env = []
  } else if (process.env.REACT_APP_INVALID_RUN_ENV !== undefined) {
    const invalidEnvs = new Set(
      process.env.REACT_APP_INVALID_RUN_ENV.split(",")
    )
    env = env.filter(e => !invalidEnvs.has(e.name))
  }

  // Set CPU value.
  let cpu: number = r && r.cpu ? r.cpu : e.cpu
  if (cpu < 512) cpu = 512

  // Set memory value.
  const memory: number = r && r.memory ? r.memory : e.memory

  // Set engine.
  const engine: ExecutionEngine = get(r, "engine", DefaultExecutionEngine)

  switch (engine) {
    case ExecutionEngine.ECS:
      return {
        cluster: get(r, "cluster", ""),
        env,
        cpu,
        memory,
        owner_id: ownerID,
        engine,
      }
    case ExecutionEngine.EKS:
    default:
      return {
        cluster: get(
          r,
          "cluster",
          process.env.REACT_APP_EKS_CLUSTER_NAME || ""
        ),
        node_lifecycle: get(r, "node_lifecycle", DefaultNodeLifecycle),
        env,
        cpu,
        memory,
        owner_id: ownerID,
        engine,
      }
  }
}
