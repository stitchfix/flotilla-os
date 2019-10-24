import { RunTaskPayload, Task, Env } from "../types"

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
}): RunTaskPayload => {
  // Set cluster value.
  const cluster = routerState && routerState.cluster ? routerState.cluster : ""

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
  return { cluster, env, cpu, memory }
}

export default getInitialValuesForTaskRun
