import getInitialValuesForTaskRun from "../getInitialValuesForTaskRun"
import { createMockTaskObject } from "../testHelpers"
import { LaunchRequestV2 } from "../../types"

describe("getInitialValuesForTaskRun", () => {
  it("works correctly", () => {
    const td = createMockTaskObject()
    const expected: LaunchRequestV2 = {
      cluster: "",
      cpu: td.cpu,
      memory: td.memory,
      env: td.env,
    }
    expect(
      getInitialValuesForTaskRun({
        task: td,
        routerState: null,
      })
    ).toEqual(expected)
  })
})
