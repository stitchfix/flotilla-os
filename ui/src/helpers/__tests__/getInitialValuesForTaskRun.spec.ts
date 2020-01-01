import getInitialValuesForTaskRun from "../getInitialValuesForTaskRun"
import { createMockTaskObject } from "../testHelpers"
import { LaunchRequestV2, ExecutionEngine } from "../../types"

describe("getInitialValuesForTaskRun", () => {
  it("works correctly", () => {
    const td = createMockTaskObject()
    const expectedEks: LaunchRequestV2 = {
      cluster: process.env.REACT_APP_DEFAULT_EXECUTION_ENGINE || "",
      cpu: td.cpu,
      memory: td.memory,
      env: td.env,
      engine: ExecutionEngine.EKS,
      command: td.command,
    }

    expect(
      getInitialValuesForTaskRun({
        task: td,
        routerState: null,
        settings: {
          USE_OPTIMIZED_LOG_RENDERER: true,
          SHOULD_OVERRIDE_CMD_F_IN_RUN_VIEW: true,
          DEFAULT_OWNER_ID: "",
        },
      })
    ).toEqual(expectedEks)
  })
})
