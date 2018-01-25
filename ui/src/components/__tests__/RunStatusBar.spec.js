import React from "react"
import { mount } from "enzyme"
import RunStatusBar from "../RunStatusBar"

jest.useFakeTimers()

describe("RunStatusBar", () => {
  it("starts the duration calculation interval when the component mounts", () => {
    const durcalc = RunStatusBar.prototype.calculateDuration
    RunStatusBar.prototype.calculateDuration = jest.fn()
    const wrapper = mount(<RunStatusBar />)

    jest.runOnlyPendingTimers()

    expect(RunStatusBar.prototype.calculateDuration).toHaveBeenCalledTimes(1)
    RunStatusBar.prototype.calculateDuration = durcalc
  })
  it("stops the duration calculation interval if the job finishes", () => {
    const stopIntervalFn = RunStatusBar.prototype.stopDurationInterval
    RunStatusBar.prototype.stopDurationInterval = jest.fn()
    const wrapper = mount(<RunStatusBar />)

    wrapper.setProps({ finishedAt: "im done" })
    expect(RunStatusBar.prototype.stopDurationInterval).toHaveBeenCalledTimes(1)
    RunStatusBar.prototype.stopDurationInterval = stopIntervalFn
  })
  it("stops the duration calculation interval if the component unmounts", () => {
    const stopIntervalFn = RunStatusBar.prototype.stopDurationInterval
    RunStatusBar.prototype.stopDurationInterval = jest.fn()
    const wrapper = mount(<RunStatusBar />)

    wrapper.unmount()
    expect(RunStatusBar.prototype.stopDurationInterval).toHaveBeenCalledTimes(1)
    RunStatusBar.prototype.stopDurationInterval = stopIntervalFn
  })
  it("renders the duration", () => {
    const fakeDuration = "some_random_string"
    const wrapper = mount(<RunStatusBar />)
    wrapper.setState({ duration: fakeDuration })

    expect(wrapper.text()).toContain(fakeDuration)
  })
  it("renders the (enhanced) status", () => {
    const wrapper = mount(<RunStatusBar />)
    expect(wrapper.find("EnhancedRunStatus")).toBeTruthy()
  })
})
