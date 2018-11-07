import React from "react"
import { mount } from "enzyme"
import EnhancedRunStatus, { getEnhancedStatus } from "../EnhancedRunStatus"
import runStatusTypes from "../../constants/runStatusTypes"

describe("getEnhancedStatus", () => {
  it("returns the correct status", () => {
    expect(getEnhancedStatus(runStatusTypes.pending)).toEqual(
      runStatusTypes.pending
    )
    expect(getEnhancedStatus(runStatusTypes.queued)).toEqual(
      runStatusTypes.queued
    )
    expect(getEnhancedStatus(runStatusTypes.running)).toEqual(
      runStatusTypes.running
    )
    expect(getEnhancedStatus(runStatusTypes.needs_retry)).toEqual(
      runStatusTypes.needs_retry
    )
    expect(getEnhancedStatus(runStatusTypes.stopped)).toEqual(
      runStatusTypes.failed
    )
    expect(getEnhancedStatus(runStatusTypes.stopped, 0)).toEqual(
      runStatusTypes.success
    )
  })
})

describe("EnhancedRunStatus", () => {
  it("returns the correct status and icon", () => {
    const pending = mount(<EnhancedRunStatus status={runStatusTypes.pending} />)
    expect(
      pending
        .find(".run-status-text")
        .text()
        .toUpperCase()
    ).toEqual(runStatusTypes.pending.toUpperCase())
    expect(pending.find("Loader").length).toEqual(1)

    const queued = mount(<EnhancedRunStatus status={runStatusTypes.queued} />)
    expect(
      queued
        .find(".run-status-text")
        .text()
        .toUpperCase()
    ).toEqual(runStatusTypes.queued.toUpperCase())
    expect(queued.find("Loader").length).toEqual(1)

    const running = mount(<EnhancedRunStatus status={runStatusTypes.running} />)
    expect(
      running
        .find(".run-status-text")
        .text()
        .toUpperCase()
    ).toEqual(runStatusTypes.running.toUpperCase())
    expect(running.find("Loader").length).toEqual(1)

    const success = mount(
      <EnhancedRunStatus status={runStatusTypes.stopped} exitCode={0} />
    )
    expect(
      success
        .find(".run-status-text")
        .text()
        .toUpperCase()
    ).toEqual(runStatusTypes.success.toUpperCase())
    expect(success.find("CheckCircle").length).toEqual(1)

    const failed = mount(
      <EnhancedRunStatus status={runStatusTypes.stopped} exitCode={1} />
    )
    expect(
      failed
        .find(".run-status-text")
        .text()
        .toUpperCase()
    ).toEqual(runStatusTypes.failed.toUpperCase())
    expect(failed.find("XCircle").length).toEqual(1)

    const needsRetry = mount(
      <EnhancedRunStatus status={runStatusTypes.needs_retry} />
    )
    expect(
      needsRetry
        .find(".run-status-text")
        .text()
        .toUpperCase()
    ).toEqual(runStatusTypes.needs_retry.toUpperCase())
    expect(needsRetry.find("XCircle").length).toEqual(1)
  })
})
