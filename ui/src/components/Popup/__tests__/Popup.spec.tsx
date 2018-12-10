import * as React from "react"
import { mount } from "enzyme"
import { UnwrappedPopup } from "../Popup"

describe("Popup", () => {
  it("renders a PopupPositioner", () => {
    const wrapper = mount(<UnwrappedPopup />)
    expect(wrapper.find("PopupPositioner").length).toBe(1)
  })

  it("renders a Card with the correct props", () => {
    const title = "title"
    const actionsID = "actionsID"
    const actions = <span id={actionsID}>actions</span>
    const wrapper = mount(<UnwrappedPopup title={title} actions={actions} />)
    const card = wrapper.find("Card")
    expect(card.length).toBe(1)
    expect(card.prop("title")).toEqual(title)
    expect(card.find(`#${actionsID}`).length).toEqual(1)
  })

  it("renders a close button", () => {
    const wrapper = mount(<UnwrappedPopup />)
    expect(wrapper.find("#popupCloseButton").length).toEqual(1)
  })

  it("calls props.unrenderPopup after specified duration if props.autohide is true", () => {
    const unrenderPopup = jest.fn(() => {})
    mount(<UnwrappedPopup unrenderPopup={unrenderPopup} />)
    expect(unrenderPopup).toHaveBeenCalledTimes(0)
    jest.runAllTimers()
    expect(unrenderPopup).toHaveBeenCalledTimes(1)
  })

  it("does not call props.unrenderPopup after specified duration if props.autohide is false", () => {
    const unrenderPopup = jest.fn(() => {})
    mount(
      <UnwrappedPopup unrenderPopup={unrenderPopup} shouldAutohide={false} />
    )
    expect(unrenderPopup).toHaveBeenCalledTimes(0)
    jest.runAllTimers()
    expect(unrenderPopup).toHaveBeenCalledTimes(0)
  })
})
