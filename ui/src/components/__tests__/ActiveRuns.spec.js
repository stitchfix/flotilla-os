import React from "react"
import queryUpdateTypes from "../../utils/queryUpdateTypes"
import { configureSetup, generateRunRes } from "../../__testutils__"
import ConnectedActiveRuns, { ActiveRuns } from "../ActiveRuns"

const baseProps = {
  clusterOptions: [],
  data: undefined,
  error: false,
  isLoading: false,
  query: {},
  updateQuery: () => {},
}

const setup = configureSetup({
  baseProps,
  connected: ConnectedActiveRuns,
  unconnected: ActiveRuns,
})

describe("ActiveRuns", () => {
  const warn = console.warn
  beforeAll(() => {
    console.warn = jest.fn()
  })
  afterAll(() => {
    console.warn = warn
  })
  it("renders 3 <SortHeaders> and 1 cluster filter", () => {
    const wrapper = setup()

    expect(wrapper.find("SortHeader").length).toBe(3)
    expect(wrapper.find("FormGroup").length).toBe(1)
    expect(wrapper.find("FormGroup").props().label).toBe("Cluster")
  })
  it("renders a loading <EmptyTable> if props.isLoading", () => {
    const wrapper = setup({
      props: { isLoading: true },
    })
    expect(wrapper.find("EmptyTable").length).toBe(1)
    expect(wrapper.find("EmptyTable").props().isLoading).toBeTruthy()
  })
  it("renders a TableError if props.error", () => {
    const err = "Uh oh."
    const wrapper = setup({
      props: { error: err },
    })

    // Expect to find the error message rendered.
    expect(wrapper.find("EmptyTable").length).toBe(1)
    expect(wrapper.find("EmptyTable").props().title).toEqual(err)
    expect(wrapper.find("EmptyTable").props().error).toBeTruthy()
    expect(wrapper.html()).toEqual(expect.stringMatching(err))
  })
  it("renders the data if present", () => {
    const wrapper = setup({
      connectToRouter: true,
      props: {
        data: {
          history: [
            // 3 active runs.
            generateRunRes("one"),
            generateRunRes("two"),
            generateRunRes("three"),
          ],
        },
      },
    })

    expect(wrapper.find("ActiveRunsRow").length).toBe(3)
    expect(wrapper.find("EmptyTable").length).toBe(0)
  })
  it("renders an <EmptyTable> if no runs are active", () => {
    const wrapper = setup({
      props: {
        data: {
          history: [],
        },
      },
    })

    expect(wrapper.find("ActiveRunsRow").length).toBe(0)
    expect(wrapper.find("EmptyTable").length).toBe(1)
  })
  it("uses props.query.cluster_name as the cluster select's value", () => {
    const clusterName = "flot1"
    const query = { cluster_name: clusterName }
    const wrapper = setup({
      connectToRouter: true,
      props: { query },
    })

    expect(wrapper.find("Select").props().value).toEqual(clusterName)
  })
  it("calls props.updateQuery when selecting a different cluster", () => {
    const newCluster = "flot2"
    const updateQuery = jest.fn()
    const wrapper = setup({
      connectToRouter: true,
      props: { updateQuery },
    })

    expect(updateQuery).toHaveBeenCalledTimes(0)

    const clusterSelect = wrapper.find("Select")
    clusterSelect.props().onChange({ value: newCluster })
    expect(updateQuery).toHaveBeenCalledTimes(1)
    expect(updateQuery).toHaveBeenCalledWith([
      {
        key: "cluster_name",
        value: newCluster,
        updateType: queryUpdateTypes.SHALLOW,
      },
      {
        key: "page",
        value: 1,
        updateType: queryUpdateTypes.SHALLOW,
      },
    ])
  })
})
