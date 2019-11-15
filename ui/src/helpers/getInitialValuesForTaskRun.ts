import { get } from "lodash"
import {
  LaunchRequestV2,
  Task,
  Env,
  ExecutionEngine,
  NodeLifecycle,
} from "../types"
import getOwnerIdRunTagFromCookies from "./getOwnerIdRunTagFromCookies"

/**
 * Given a task definition and history.location.state, return the initial
 * values for the RunForm.
 */
const getInitialValuesForTaskRun = ({
  task,
  routerState,
}: {
  task: Task
  routerState: any
}): LaunchRequestV2 => {
  // Set ownerID value.
  const ownerID = get(
    routerState,
    ["run_tags", "owner_id"],
    getOwnerIdRunTagFromCookies()
  )

  // Set env value.
  let env: Env[] = routerState && routerState.env ? routerState.env : task.env

  // Filter out invalid run env if specified in dotenv file.
  if (process.env.REACT_APP_INVALID_RUN_ENV !== undefined) {
    const invalidEnvs = new Set(
      process.env.REACT_APP_INVALID_RUN_ENV.split(",")
    )
    env = env.filter(e => !invalidEnvs.has(e.name))
  }

  // Set CPU value.
  let cpu: number = routerState && routerState.cpu ? routerState.cpu : task.cpu
  if (cpu < 512) cpu = 512

  // Set memory value.
  const memory: number =
    routerState && routerState.memory ? routerState.memory : task.memory

  // Set command value.
  const command: string =
    routerState && routerState.command ? routerState.command : task.command

  // Set engine.
  const engine: ExecutionEngine = get(
    routerState,
    "engine",
    process.env.REACT_APP_DEFAULT_EXECUTION_ENGINE
  )

  switch (engine) {
    case ExecutionEngine.ECS:
      return {
        cluster: get(routerState, "cluster", ""),
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
          routerState,
          "cluster",
          process.env.REACT_APP_EKS_CLUSTER_NAME
        ),
        node_lifecycle: get(
          routerState,
          "node_lifecycle",
          process.env.REACT_APP_DEFAULT_NODE_LIFECYCLE
        ),
        ephemeral_storage: get(routerState, "ephemeral_storage", null),
        env,
        cpu,
        memory,
        owner_id: ownerID,
        engine,
        command,
      }
  }
}

export default getInitialValuesForTaskRun
