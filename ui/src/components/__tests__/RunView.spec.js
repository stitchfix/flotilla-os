import React from "react"
import qs from "query-string"
import { configureSetup, generateRunRes } from "../../__testutils__"
import { runStatusTypes } from "../../constants/"
import { getRetryEnv } from "../../utils/"
import { RunView } from "../RunView"

const runId = "some_id"
const data = generateRunRes(runId)
const setup = configureSetup({
  unconnected: RunView,
  baseProps: { data, runId },
})

describe("RunView", () => {
  it("renders a View component", () => {
    const wrapper = setup({ connectToRouter: true })
    expect(wrapper.find("View").length).toEqual(1)
  })
  it("renders a ViewHeader component with the correct props", () => {
    const wrapper = setup({ connectToRouter: true })
    const vh = wrapper.find("ViewHeader")
    expect(vh.length).toEqual(1)
    expect(vh.text()).toMatch(runId)

    const retryLink = vh.find("Link")
    expect(retryLink.length).toEqual(1)
    expect(retryLink.props().to).toEqual({
      pathname: `/tasks/${data.definition_id}/run`,
      search: `?cluster=${data.cluster}&${qs.stringify({
        env: getRetryEnv(data.env),
      })}`,
    })

    expect(vh.find("Button").length).toEqual(1)
  })
  // @TODO
  it("renders a RunInfo component", () => {
    const wrapper = setup({ connectToRouter: true })
    expect(wrapper.find("RunInfo").length).toEqual(1)
  })
  it("renders a RunLogs component", () => {
    const wrapper = setup({ connectToRouter: true })
    expect(wrapper.find("RunLogs").length).toEqual(1)
  })
})
