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
  const cluster = routerState && routerState.cluster ? routerState.cluster : ""
  let env: Env[] = routerState && routerState.env ? routerState.env : task.env

  // Filter out invalid run env.
  if (process.env.REACT_APP_INVALID_RUN_ENV !== undefined) {
    const invalidEnvs = new Set(
      process.env.REACT_APP_INVALID_RUN_ENV.split(",")
    )
    env = env.filter(e => !invalidEnvs.has(e.name))
  }

  const cpu: number =
    routerState && routerState.cpu ? routerState.cpu : task.cpu
  const memory: number =
    routerState && routerState.memory ? routerState.memory : task.memory
  return { cluster, env, cpu, memory }
}

export default getInitialValuesForTaskRun
