import React from "react"
import qs from "qs"
import { configureSetup, generateRunRes } from "../../__testutils__"
import runStatusTypes from "../../constants/runStatusTypes"
import getRetryEnv from "../../utils/getRetryEnv"
import { RunView } from "../RunView"

const runId = "some_id"
const data = generateRunRes(runId)
const setup = configureSetup({
  unconnected: RunView,
  baseProps: { data, runId },
})

describe("RunView", () => {
  const warn = console.warn
  const error = console.error
  beforeAll(() => {
    console.warn = jest.fn()
    console.error = jest.fn()
  })
  afterAll(() => {
    console.warn = warn
    console.error = error
  })
  it("renders a View component", () => {
    const wrapper = setup({ connectToRouter: true })
    expect(wrapper.find(".pl-view-container").length).toEqual(1)
    expect(wrapper.find(".pl-view-inner").length).toEqual(1)
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
  it("renders a RunInfo component", () => {
    const wrapper = setup({ connectToRouter: true })
    expect(wrapper.find("RunInfo").length).toEqual(1)
  })
  it("renders a RunLogs component", () => {
    const wrapper = setup({ connectToRouter: true })
    expect(wrapper.find("RunLogs").length).toEqual(1)
  })
})
