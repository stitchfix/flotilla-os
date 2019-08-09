import * as React from "react"
import { mount } from "enzyme"
import flushPromiseQueue from "flush-promises"
import Logs from "../Logs"
import { RunLog, RunStatus } from "../../types"
import { RequestStatus } from "../Request"

describe("Logs", () => {
  it("sets state correctly during the request process", async () => {
    jest.useFakeTimers()
    const logChunks: RunLog[] = [
      { log: "a", last_seen: "a" },
      { log: "a", last_seen: "a" },
      { log: "b", last_seen: "b" },
      { log: "c", last_seen: "c" },
      { log: "d", last_seen: "d" },
    ]
    let count = 0
    const requestFn = jest.fn(
      () =>
        new Promise<RunLog>(resolve => {
          resolve(logChunks[count])
          count += 1
        })
    )
    const runID = "run_one"

    expect(requestFn).toHaveBeenCalledTimes(0)

    // Mount the component.
    const wrapper = mount<Logs>(
      <Logs runID={runID} status={RunStatus.RUNNING} requestFn={requestFn} />
    )

    // `requestFn` should be called once when the component mounts.
    // Additionally, an interval will be set to call `requestFn` every n
    // seconds.
    expect(requestFn).toHaveBeenCalledTimes(1)
    expect(wrapper.state()).toEqual({
      requestStatus: RequestStatus.NOT_READY,
      data: [],
      isLoading: true,
      error: null,
      lastSeen: undefined,
      totalLogsLength: 0,
    })

    // Flush promises.
    await flushPromiseQueue()

    // The logs should be added to the component's state.
    expect(wrapper.state()).toEqual({
      requestStatus: RequestStatus.READY,
      data: [logChunks[0]],
      isLoading: false,
      error: null,
      lastSeen: logChunks[0].last_seen,
      totalLogsLength: logChunks[0].log.length,
    })

    // Advance timer, flush promises.
    jest.runOnlyPendingTimers()
    await flushPromiseQueue()

    // `requestFn` should be called again. This time, however, since the next
    // log chunk's last seen key is equal to this.state.lastSeen, we do not
    // append the logs to state.
    expect(requestFn).toHaveBeenCalledTimes(2)
    expect(wrapper.state()).toEqual({
      requestStatus: RequestStatus.READY,
      data: [logChunks[0]],
      isLoading: false,
      error: null,
      lastSeen: logChunks[0].last_seen,
      totalLogsLength: logChunks[0].log.length,
    })

    // Advance timer, flush promises.
    jest.runOnlyPendingTimers()
    await flushPromiseQueue()

    // `requestFn` should be called again. This time, since the next log
    // chunk's (logChunks[2]) `last_seen` property is different, we append the
    // logs to state.
    expect(requestFn).toHaveBeenCalledTimes(3)
    expect(wrapper.state()).toEqual({
      requestStatus: RequestStatus.READY,
      data: [logChunks[0], logChunks[2]],
      isLoading: false,
      error: null,
      lastSeen: logChunks[2].last_seen,
      totalLogsLength: logChunks[0].log.length + logChunks[2].log.length,
    })

    expect(wrapper.instance().requestInterval).toEqual(expect.any(Number))
    // Simulate the run going from a `running` state to a `stopped` state.
    // The component's `requestInterval` member should now be cleared and we
    // will "exhaust" the logs endpoint until all logs have been received.
    wrapper.setProps({ status: RunStatus.STOPPED })
    expect(wrapper.instance().requestInterval).toEqual(undefined)

    // `requestFn` should be called once more to ensure that all logs have
    // been fetched.
    expect(requestFn).toHaveBeenCalledTimes(4)

    // Flush promises
    await flushPromiseQueue()

    // We should hit the API two more times - once to get logChunks[4] and
    // another time to ensure that there are no more logs.
    expect(requestFn).toHaveBeenCalledTimes(6)
    expect(wrapper.state()).toEqual({
      requestStatus: RequestStatus.READY,
      data: [logChunks[0], logChunks[2], logChunks[3], logChunks[4]],
      isLoading: false,
      error: null,
      lastSeen: logChunks[4].last_seen,
      totalLogsLength:
        logChunks[0].log.length +
        logChunks[2].log.length +
        logChunks[3].log.length +
        logChunks[4].log.length,
    })
  })
})
