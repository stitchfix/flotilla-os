import * as React from "react"
import { mount } from "enzyme"
import { UnwrappedConfirmModal } from "../ConfirmModal"

describe("ConfirmModal", () => {
  beforeAll(() => {})
  it("renders a cancel button", () => {
    const requestFn = jest.fn()
    const unrenderModal = jest.fn()
    const wrapper = mount(
      <UnwrappedConfirmModal
        requestFn={requestFn}
        unrenderModal={unrenderModal}
      />
    )
    const cancelButton = wrapper.find("Button").at(0)
    expect(cancelButton.is("#cancel")).toEqual(true)
    expect(unrenderModal).toHaveBeenCalledTimes(0)
    cancelButton.simulate("click")
    expect(unrenderModal).toHaveBeenCalledTimes(1)
  })
  it("renders a confirm button", async () => {
    const requestArgs = { foo: "bar" }
    const requestFn = jest.fn(() => Promise.resolve())
    const wrapper = mount(
      <UnwrappedConfirmModal
        renderPopup={() => {}}
        unrenderPopup={() => {}}
        unrenderModal={() => {}}
        requestFn={requestFn}
        getRequestArgs={() => requestArgs}
      />
    )
    const confirmButton = wrapper.find("Button").at(1)
    expect(confirmButton.is("#confirm")).toEqual(true)
    expect(requestFn).toHaveBeenCalledTimes(0)
    confirmButton.simulate("click")
    expect(requestFn).toHaveBeenCalledTimes(1)
    expect(requestFn).toHaveBeenCalledWith(requestArgs)
  })
})
