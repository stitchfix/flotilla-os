import * as React from "react"
import { mount } from "enzyme"
import PopupContainer from "../PopupContainer"
import { IPopupProps } from "../../../.."

describe("PopupContainer", () => {
  it("can render and unrender popups", () => {
    const wrapper = mount(
      <PopupContainer>
        <div>test</div>
      </PopupContainer>
    )

    expect(wrapper.state("isVisible")).toEqual(false)
    expect(wrapper.state("popupProps")).toEqual(undefined)

    // Access instance methods per: https://github.com/airbnb/enzyme/issues/208#issuecomment-344401247
    const instance = wrapper.instance() as PopupContainer
    const popupProps: IPopupProps = {
      shouldAutohide: true,
      unrenderPopup: () => {},
      visibleDuration: 5000,
    }
    instance.renderPopup(popupProps)

    expect(wrapper.state("isVisible")).toEqual(true)
    expect(wrapper.state("popupProps")).toEqual(popupProps)

    instance.unrenderPopup()
    expect(wrapper.state("isVisible")).toEqual(false)
    expect(wrapper.state("popupProps")).toEqual(undefined)
  })
})
