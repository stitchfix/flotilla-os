import React from "react"
import { mount } from "enzyme"
import { MemoryRouter } from "react-router-dom"
import { configureSetup, generateRunRes } from "../../__testutils__"
import ActiveRunsRow from "../ActiveRunsRow"

const runId = "runId"
const setup = configureSetup({
  baseProps: {
    data: generateRunRes(runId),
  },
  unconnected: ActiveRunsRow,
})

describe("ActiveRunsRow", () => {
  const warn = console.warn
  beforeAll(() => {
    console.warn = jest.fn()
  })
  afterAll(() => {
    console.warn = warn
  })
  it("renders a Link to /runs/:run_id", () => {
    const wrapper = setup({
      connectToRouter: true,
    })
    expect(wrapper.find("Link").length).toBe(1)
    expect(wrapper.find("Link").props().to).toEqual(`/runs/${runId}`)
  })
  it("calls props.onStopButtonClick when clicking on the Stop Button", () => {
    const onStopButtonClick = jest.fn()
    const wrapper = setup({
      connectToRouter: true,
      props: { onStopButtonClick },
    })
    wrapper.find("Button").simulate("click")
    expect(onStopButtonClick).toHaveBeenCalledTimes(1)
  })
})
