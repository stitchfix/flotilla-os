import * as React from "react"
import { shallow } from "enzyme"
import { LogProcessor } from "../LogProcessor"

jest.mock("../../workers/index")

describe("LogProcessor", () => {
  it("calls processLogs upon mounting and if logs/width changes", () => {
    const process = LogProcessor.prototype.processLogs
    LogProcessor.prototype.processLogs = jest.fn()
    const wrapper = shallow(
      <LogProcessor logs="abc" width={100} height={100} />
    )
    expect(LogProcessor.prototype.processLogs).toHaveBeenCalledTimes(1)
    wrapper.setProps({ logs: "abcdefg" })
    expect(LogProcessor.prototype.processLogs).toHaveBeenCalledTimes(2)
    LogProcessor.prototype.processLogs = process
  })
})
