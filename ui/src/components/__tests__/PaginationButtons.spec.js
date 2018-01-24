import React from "react"
import { configureSetup } from "../../__testutils__"
import PaginationButtons from "../PaginationButtons"

const buttonCount = 5
const limit = 20
const className = "pagination-button"
const buttonEl = <button className={className} />
const baseProps = {
  total: 100,
  buttonCount,
  offset: 0,
  limit,
  updateQuery: () => {},
  buttonEl,
}

const setup = configureSetup({
  unconnected: PaginationButtons,
  baseProps,
})

describe("PaginationButtons", () => {
  it("renders the correct number of buttons", () => {
    // If total pages >= the number of buttons to render, render the exact
    // number of buttons specified, e.g. props.buttonCount
    const wrapper = setup()
    expect(wrapper.find(`.${className}`).length).toBe(buttonCount)

    // If total pages (in this case, 2) < number of buttons to render (5),
    // should only render 2 buttons.
    const total = 40
    wrapper.setProps({ total })
    expect(wrapper.find(`.${className}`).length).toBe(total / limit)
  })
  it("adds props.activeButtonClassName to the active button", () => {
    const wrapper = setup()

    // Initially, the first button should be active since offset is 0
    expect(
      wrapper
        .find(`.${className}`)
        .at(0)
        .hasClass(wrapper.props().activeButtonClassName)
    ).toEqual(true)
    expect(
      wrapper
        .find(`.${className}`)
        .at(1)
        .hasClass(wrapper.props().activeButtonClassName)
    ).toEqual(false)

    // Increase the offset by limit * 2, the third button should now be active
    wrapper.setProps({ offset: limit * 2 })
    expect(
      wrapper
        .find(`.${className}`)
        .at(0)
        .hasClass(wrapper.props().activeButtonClassName)
    ).toEqual(false)
    expect(
      wrapper
        .find(`.${className}`)
        .at(1)
        .hasClass(wrapper.props().activeButtonClassName)
    ).toEqual(false)
    expect(
      wrapper
        .find(`.${className}`)
        .at(2)
        .hasClass(wrapper.props().activeButtonClassName)
    ).toEqual(true)
  })
  it("adds an onClick prop to each button", () => {
    const updateQuery = jest.fn()
    const wrapper = setup({
      props: { updateQuery },
    })

    wrapper.find(`.${className}`).forEach(btn => {
      expect(typeof btn.props().onClick).toEqual("function")
    })

    // Each onClick function should call props.updateQuery with the correct
    // page number
    const desiredPage = 3
    const buttonIndexOfDesiredPage = desiredPage - 1
    wrapper
      .find(`.${className}`)
      .at(buttonIndexOfDesiredPage)
      .simulate("click")
    expect(updateQuery).toHaveBeenCalledWith({
      key: "page",
      value: desiredPage,
      updateType: "SHALLOW",
    })
  })
})
