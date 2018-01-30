import React from "react"
import { mount } from "enzyme"
import { MemoryRouter } from "react-router-dom"
import { configureSetup, generateTaskRes } from "../../__testutils__"
import ConnectedTasks, { Tasks } from "../Tasks"

const baseProps = {
  isLoading: false,
  error: false,
  data: undefined,
  updateQuery: () => {},
  query: {},
  clusterOptions: [],
  dispatch: () => {},
}
const setup = configureSetup({
  connected: ConnectedTasks,
  unconnected: Tasks,
  baseProps,
})

describe("Tasks", () => {
  let warn = console.warn
  let error = console.error
  beforeAll(() => {
    console.warn = jest.fn()
    console.error = jest.fn()
  })
  afterAll(() => {
    console.warn = warn
    console.error = error
  })
  it("renders 3 SortHeaders", () => {
    const wrapper = setup({ connectToRouter: true })

    expect(wrapper.find("SortHeader").length).toBe(3)
  })
  it("renders 3 filters", () => {
    const wrapper = setup({
      connectToRouter: true,
    })

    expect(wrapper.find("FormGroup").length).toEqual(3)
    expect(
      wrapper
        .find("FormGroup")
        .at(0)
        .props().label
    ).toEqual("Alias")
    expect(
      wrapper
        .find("FormGroup")
        .at(1)
        .props().label
    ).toEqual("Group Name")
    expect(
      wrapper
        .find("FormGroup")
        .at(2)
        .props().label
    ).toEqual("Image")
  })
  it("renders a <Loader> if props.isLoading", () => {
    const wrapper = setup({
      props: { isLoading: true },
      connectToRouter: true,
    })
    expect(wrapper.find("Loader").length).toBe(1)
  })
  it("renders a TableError if props.error", () => {
    const err = "Uh oh."
    const wrapper = setup({
      props: { error: err },
      connectToRouter: true,
    })

    // Expect to find the error message rendered.
    expect(wrapper.find("EmptyTable").length).toBe(1)
    expect(wrapper.find("EmptyTable").props().title).toEqual(err)
    expect(wrapper.find("EmptyTable").props().error).toBeTruthy()
    expect(wrapper.html()).toEqual(expect.stringMatching(err))
    expect(wrapper.find("Loader").length).toBe(0)
  })
  it("renders the data if present", () => {
    const wrapper = setup({
      props: {
        data: {
          definitions: [
            generateTaskRes("one"),
            generateTaskRes("two"),
            generateTaskRes("three"),
          ],
        },
      },
      connectToRouter: true,
    })

    expect(wrapper.find("TasksRow").length).toBe(3)
  })
})
