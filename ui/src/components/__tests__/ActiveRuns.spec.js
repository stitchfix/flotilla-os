import React from "react"
import { mount } from "enzyme"
import { MemoryRouter } from "react-router-dom"
import { generateRunRes } from "../../__testutils__"
import ConnectedActiveRuns, { ActiveRuns } from "../ActiveRuns"

const baseProps = {
  isLoading: false,
  error: false,
  data: undefined,
  updateQuery: () => {},
  query: {},
  clusterOptions: [],
  dispatch: () => {},
}
const setup = (props = {}) => {
  const mergedProps = {
    ...baseProps,
    ...props,
  }
  return mount(
    <MemoryRouter>
      <ActiveRuns {...mergedProps} />
    </MemoryRouter>
  )
}

describe("ActiveRuns", () => {
  it("renders 3 <SortHeaders> and 1 cluster filter", () => {
    const wrapper = setup()

    expect(wrapper.find("SortHeader").length).toBe(3)
    expect(wrapper.find("FormGroup").length).toBe(1)
    expect(wrapper.find("FormGroup").props().label).toBe("Cluster")
  })
  it("renders a <Loader> if props.isLoading", () => {
    const wrapper = setup({ isLoading: true })
    expect(wrapper.find("Loader").length).toBe(1)
  })
  it("renders a TableError if props.error", () => {
    const err = "Uh oh."
    const wrapper = setup({ error: err })

    // Expect to find the error message rendered.
    expect(wrapper.find(".table-error-container").length).toBe(1)
    expect(wrapper.html()).toEqual(expect.stringMatching(err))
    expect(wrapper.find("Loader").length).toBe(0)
  })
  it("renders the data if present", () => {
    const wrapper = setup({
      data: {
        history: [
          // 3 active runs.
          generateRunRes("one"),
          generateRunRes("two"),
          generateRunRes("three"),
        ],
      },
    })

    expect(wrapper.find("ActiveRunsRow").length).toBe(3)
    expect(wrapper.find(".table-error-container").length).toBe(0)
    expect(wrapper.find("Loader").length).toBe(0)
  })
})
