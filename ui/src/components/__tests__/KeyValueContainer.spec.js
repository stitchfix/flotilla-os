import React from "react"
import { configureSetup } from "../../__testutils__"
import KeyValueContainer from "../KeyValueContainer"

const header = "header"
const children = () => {}
const setup = configureSetup({
  baseProps: { header, children },
  unconnected: KeyValueContainer,
})

describe("KeyValueContainer", () => {
  let wrapper
  beforeEach(() => {
    wrapper = setup()
  })
  it("renders the appropriate title and buttons in the header", () => {
    // Ensure that initial state is what we expect.
    expect(wrapper.state().json).toBeFalsy()
    expect(wrapper.state().collapsed).toBeFalsy()

    // Non-JSON, non-collapsed.
    expect(
      wrapper
        .find("Button")
        .at(0)
        .text()
    ).toEqual("JSON View")
    expect(wrapper.find("ChevronUp").length).toBe(1)
    expect(wrapper.find("ChevronDown").length).toBe(0)

    // JSON, non-collapsed.
    wrapper.setState({ json: true })
    expect(
      wrapper
        .find("Button")
        .at(0)
        .text()
    ).toEqual("Normal View")
    expect(wrapper.find("ChevronUp").length).toBe(1)
    expect(wrapper.find("ChevronDown").length).toBe(0)

    // JSON, collapsed.
    wrapper.setState({ collapsed: true })
    expect(
      wrapper
        .find("Button")
        .at(0)
        .text()
    ).toEqual("Normal View")
    expect(wrapper.find("ChevronUp").length).toBe(0)
    expect(wrapper.find("ChevronDown").length).toBe(1)

    // Non-JSON, collapsed
    wrapper.setState({ json: false })
    expect(
      wrapper
        .find("Button")
        .at(0)
        .text()
    ).toEqual("JSON View")
    expect(wrapper.find("ChevronUp").length).toBe(0)
    expect(wrapper.find("ChevronDown").length).toBe(1)
  })
  it("handles the JSON-toggle button click", () => {
    // Ensure that initial state is what we expect.
    expect(wrapper.state().json).toBeFalsy()
    expect(wrapper.state().collapsed).toBeFalsy()

    // Click it once to make it JSON.
    wrapper.instance().handleJsonButtonClick()

    expect(wrapper.state().json).toBeTruthy()
    expect(wrapper.state().collapsed).toBeFalsy()

    // Manually set the collapsed state to true, then call
    // handleJsonButtonClick. `collapsed` should now be false.
    wrapper.setState({ collapsed: true })
    wrapper.instance().handleJsonButtonClick()

    expect(wrapper.state().json).toBeFalsy()
    expect(wrapper.state().collapsed).toBeFalsy()
  })
  it("handles the collapsed-toggle button click", () => {
    // Ensure that initial state is what we expect.
    expect(wrapper.state().json).toBeFalsy()
    expect(wrapper.state().collapsed).toBeFalsy()

    // Click it once.
    wrapper.instance().handleCollapseButtonClick()
    expect(wrapper.state().json).toBeFalsy()
    expect(wrapper.state().collapsed).toBeTruthy()

    // Click it again.
    wrapper.instance().handleCollapseButtonClick()
    expect(wrapper.state().json).toBeFalsy()
    expect(wrapper.state().collapsed).toBeFalsy()
  })
  it("returns the appropriate state to it's children", () => {
    expect(wrapper.instance().getState()).toEqual(wrapper.state())
  })
})
