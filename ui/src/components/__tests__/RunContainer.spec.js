import React from "react"
import { shallow } from "enzyme"
import fetchMock from "fetch-mock"
import { RunContainer } from "../RunContainer"

jest.useFakeTimers()

const runId = "asdf"
const res = { foo: "bar" }
fetchMock.get("*", res)

describe("RunContainer", () => {
  describe("Lifecycle Methods", () => {
    let fetch = RunContainer.prototype.fetch
    let startInterval = RunContainer.prototype.startInterval
    let stopInterval = RunContainer.prototype.stopInterval
    beforeEach(() => {
      RunContainer.prototype.fetch = jest.fn()
      RunContainer.prototype.startInterval = jest.fn()
      RunContainer.prototype.stopInterval = jest.fn()
    })
    afterEach(() => {
      RunContainer.prototype.fetch = fetch
      RunContainer.prototype.startInterval = startInterval
      RunContainer.prototype.stopInterval = stopInterval
    })
    describe("componentDidMount", () => {
      it("calls props.fetch with the router's `runId` param and starts the fetch interval", () => {
        const wrapper = shallow(<RunContainer match={{ params: { runId } }} />)
        expect(RunContainer.prototype.fetch).toHaveBeenCalledTimes(1)
        expect(RunContainer.prototype.fetch).toHaveBeenCalledWith(
          expect.stringContaining(`/task/history/${runId}`)
        )
        expect(RunContainer.prototype.startInterval).toHaveBeenCalledTimes(1)
      })
    })
    describe("componentWillReceiveProps", () => {
      it("calls props.fetch with nextProps.match.params.runId when the runId changes", () => {
        const wrapper = shallow(<RunContainer match={{ params: { runId } }} />)
        expect(RunContainer.prototype.fetch).toHaveBeenCalledTimes(1)
        expect(RunContainer.prototype.startInterval).toHaveBeenCalledTimes(1)

        const nextRunId = "nextRunId"
        wrapper.setProps({ match: { params: { runId: nextRunId } } })
        expect(RunContainer.prototype.fetch).toHaveBeenCalledTimes(2)
        expect(RunContainer.prototype.fetch).toHaveBeenCalledWith(
          expect.stringContaining(`/task/history/${nextRunId}`)
        )
        expect(RunContainer.prototype.stopInterval).toHaveBeenCalledTimes(1)
        expect(RunContainer.prototype.startInterval).toHaveBeenCalledTimes(2)
      })
      it("stops the interval if the run has stopped", () => {
        const wrapper = shallow(<RunContainer match={{ params: { runId } }} />)
        wrapper.setProps({ data: { status: "STOPPED" } })
        expect(RunContainer.prototype.stopInterval).toHaveBeenCalledTimes(1)
      })
    })
  })
})
