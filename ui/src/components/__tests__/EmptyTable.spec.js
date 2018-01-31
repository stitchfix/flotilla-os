import React from "react"
import { configureSetup } from "../../__testutils__"
import EmptyTable from "../EmptyTable"

const rootDivClassName = "flot-empty-table"
const title = "Something happened!"
const actions = "Do something?"
const setup = configureSetup({
  baseProps: { title, actions },
  unconnected: EmptyTable,
})

describe("EmptyTable", () => {
  it("renders a <Loader> if props.isLoading", () => {
    const wrapper = setup({
      props: { isLoading: true },
    })
    expect(wrapper.find("Loader").length).toBe(1)
    expect(wrapper.find(`.${rootDivClassName}`).length).toBe(0)
  })
  it("adds an `error` class to the root div if props.error", () => {
    const wrapper = setup({
      props: { error: true },
    })
    expect(wrapper.find(`.${rootDivClassName}.error`).length).toBe(1)
  })
  it("renders props.title and props.actions", () => {
    const wrapper = setup()
    expect(wrapper.find("Loader").length).toBe(0)
    expect(wrapper.find(`.${rootDivClassName}`).length).toBe(1)
    expect(wrapper.find(`.flot-empty-table-title`).length).toBe(1)
    expect(wrapper.find(`.flot-empty-table-title`).text()).toBe(title)
    expect(wrapper.find(`.flot-empty-table-actions`).length).toBe(1)
    expect(wrapper.find(`.flot-empty-table-actions`).text()).toBe(actions)
  })
})
