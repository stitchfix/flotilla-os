import React from "react"
import { shallow } from "enzyme"
import axios from "axios"
import axiosMockAdapter from "axios-mock-adapter"
import RunLogs from "../RunLogs"
import { simpleLogRes, configureSetup } from "../../__testutils__/"
import runStatusTypes from "../../constants/runStatusTypes"
import config from "../../config"

jest.useFakeTimers()

const axiosMock = new axiosMockAdapter(axios)
const runId = "runId"
const baseProps = {
  runId,
  status: runStatusTypes.running,
}
const setup = configureSetup({
  unconnected: RunLogs,
  baseProps,
})

describe("RunLogs", () => {
  describe("Lifecycle Methods", () => {
    let realFetch = RunLogs.prototype.fetch
    let realStartInterval = RunLogs.prototype.startInterval
    let realStopInterval = RunLogs.prototype.stopInterval
    beforeEach(() => {
      RunLogs.prototype.fetch = jest.fn()
      RunLogs.prototype.startInterval = jest.fn()
      RunLogs.prototype.stopInterval = jest.fn()
    })
    afterEach(() => {
      RunLogs.prototype.fetch = realFetch
      RunLogs.prototype.startInterval = realStartInterval
      RunLogs.prototype.stopInterval = realStopInterval
    })
    describe("componentDidMount", () => {
      it("calls this.fetch with the router's `runId` param and calls this.startInterval", () => {
        setup()
        expect(RunLogs.prototype.fetch).toHaveBeenCalledTimes(1)
        expect(RunLogs.prototype.fetch).toHaveBeenCalledWith(runId)
        expect(RunLogs.prototype.startInterval).toHaveBeenCalledTimes(1)
      })
    })
    describe("componentWillReceiveProps", () => {
      it("calls props.fetch with nextProps.match.params.runId when the runId changes", () => {
        const wrapper = setup()
        expect(RunLogs.prototype.fetch).toHaveBeenCalledTimes(1)
        expect(RunLogs.prototype.startInterval).toHaveBeenCalledTimes(1)

        const nextRunId = "nextRunId"
        wrapper.setProps({ runId: nextRunId })
        expect(RunLogs.prototype.fetch).toHaveBeenCalledTimes(2)
        expect(RunLogs.prototype.fetch).toHaveBeenCalledWith(nextRunId)
        expect(RunLogs.prototype.stopInterval).toHaveBeenCalledTimes(1)
        expect(RunLogs.prototype.startInterval).toHaveBeenCalledTimes(2)
      })
      it("stops the interval if the run has stopped", () => {
        const wrapper = setup()
        wrapper.setProps({ status: "STOPPED" })
        expect(RunLogs.prototype.stopInterval).toHaveBeenCalledTimes(1)
      })
    })
    describe("componentWillUnmount", () => {
      it("calls this.stopInterval", () => {
        const wrapper = setup()
        wrapper.unmount()
        expect(RunLogs.prototype.stopInterval).toHaveBeenCalledTimes(1)
      })
    })
  })
  describe("Non-lifecycle Methods", () => {
    let cdm = RunLogs.prototype.componentDidMount
    let cwrp = RunLogs.prototype.componentWillReceiveProps
    let cwu = RunLogs.prototype.componentWillUnmount
    beforeEach(() => {
      RunLogs.prototype.componentDidMount = jest.fn()
      RunLogs.prototype.componentWillReceiveProps = jest.fn()
      RunLogs.prototype.componentWillUnmount = jest.fn()
    })
    afterEach(() => {
      RunLogs.prototype.componentDidMount = cdm
      RunLogs.prototype.componentWillReceiveProps = cwrp
      RunLogs.prototype.componentWillUnmount = cwu
    })
    describe("fetch", () => {
      it("appends logs to state", async () => {
        axiosMock
          .onGet(`${config.FLOTILLA_API}/${runId}/logs`)
          .reply(200, simpleLogRes)

        const wrapper = setup()

        await wrapper.instance().fetch(runId)
        expect(wrapper.state().logs).toEqual(simpleLogRes.log.split("\n"))
      })
    })
    describe("startInterval", () => {
      it("starts an interval and sets state.isLoading to true", () => {
        let fetch = RunLogs.prototype.fetch
        RunLogs.prototype.fetch = jest.fn()
        const wrapper = setup()

        wrapper.instance().startInterval()

        jest.runOnlyPendingTimers()
        expect(RunLogs.prototype.fetch).toHaveBeenCalledTimes(1)

        jest.runOnlyPendingTimers()
        expect(RunLogs.prototype.fetch).toHaveBeenCalledTimes(2)
        expect(wrapper.state().isLoading).toBe(true)

        RunLogs.prototype.fetch = fetch
      })
    })
    describe("stopInterval", () => {
      it("stops the interval and sets state.isLoading to false", () => {
        let fetch = RunLogs.prototype.fetch
        RunLogs.prototype.fetch = jest.fn()
        const wrapper = setup()

        wrapper.instance().startInterval()
        expect(wrapper.state().isLoading).toBe(true)

        wrapper.instance().stopInterval()
        expect(wrapper.state().isLoading).toBe(false)

        RunLogs.prototype.fetch = fetch
      })
    })
  })
})
