import getInitialValuesForTaskRun from "../getInitialValuesForTaskRun"
import { createMockTaskObject } from "../testHelpers"
import { RunTaskPayload } from "../../types"

describe("getInitialValuesForTaskRun", () => {
  it("works correctly", () => {
    const td = createMockTaskObject()
    const expected: RunTaskPayload = {
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
