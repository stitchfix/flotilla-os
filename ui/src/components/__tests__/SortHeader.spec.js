import React from "react"
import { configureSetup } from "../../__testutils__"
import SortHeader from "../SortHeader"

const sk1 = "sk1"
const sk2 = "sk2"
const baseProps = {
  currentSortKey: null,
  currentOrder: null,
  display: "hi",
  sortKey: sk1,
}
const setup = configureSetup({
  unconnected: SortHeader,
  baseProps,
})

describe("SortHeader", () => {
  it("renders a button with the correct classNames", async () => {
    const wrapper = setup()

    expect(wrapper.find("button").length).toEqual(1)
    expect(wrapper.find("button").hasClass("pl-th")).toEqual(true)
    expect(wrapper.find("button").hasClass("pl-th-sort")).toEqual(true)

    // Set currentSortKey to sortKey
    wrapper.setProps({ currentSortKey: sk1, currentOrder: "asc" })

    expect(wrapper.find("button").hasClass("pl-th-sort-active")).toEqual(true)
    expect(wrapper.find("button").hasClass("asc")).toEqual(true)
    expect(wrapper.find("button").hasClass("desc")).toEqual(false)

    // Change order
    wrapper.setProps({ currentOrder: "desc" })

    expect(wrapper.find("button").hasClass("pl-th-sort-active")).toEqual(true)
    expect(wrapper.find("button").hasClass("asc")).toEqual(false)
    expect(wrapper.find("button").hasClass("desc")).toEqual(true)

    // Set to sk2
    wrapper.setProps({ currentSortKey: sk2, currentOrder: "asc" })
    expect(wrapper.find("button").hasClass("pl-th-sort-active")).toEqual(false)
  })
  it("calculates the correct next sort state and calls props.updateQuery with the correct data structure", () => {
    const updateQuery = jest.fn()
    const wrapper = setup({
      props: { updateQuery },
    })
    const button = wrapper.find("button")

    button.simulate("click")
    expect(updateQuery).toHaveBeenCalledWith([
      {
        key: "sort_by",
        value: sk1,
        updateType: "SHALLOW",
      },
      {
        key: "order",
        value: "asc",
        updateType: "SHALLOW",
      },
      {
        key: "page",
        value: 1,
        updateType: "SHALLOW",
      },
    ])

    wrapper.setProps({
      currentOrder: "asc",
      currentSortKey: sk1,
    })
    button.simulate("click")
    expect(updateQuery).toHaveBeenCalledWith([
      {
        key: "sort_by",
        value: sk1,
        updateType: "SHALLOW",
      },
      {
        key: "order",
        value: "desc",
        updateType: "SHALLOW",
      },
      {
        key: "page",
        value: 1,
        updateType: "SHALLOW",
      },
    ])

    wrapper.setProps({
      currentOrder: "desc",
      currentSortKey: sk1,
    })
    button.simulate("click")
    expect(updateQuery).toHaveBeenCalledWith([
      {
        key: "sort_by",
        value: null,
        updateType: "SHALLOW",
      },
      {
        key: "order",
        value: null,
        updateType: "SHALLOW",
      },
      {
        key: "page",
        value: 1,
        updateType: "SHALLOW",
      },
    ])
  })
})
