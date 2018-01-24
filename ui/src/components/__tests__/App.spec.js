import React from "react"
import { mount } from "enzyme"
import { MemoryRouter } from "react-router"
import { App } from "../App"

const baseProps = {
  modal: {
    modalVisible: false,
    modal: undefined,
  },
  popup: {
    popupVisible: false,
    popup: undefined,
  },
}

const setup = (props = {}) => {
  const mergedProps = {
    ...baseProps,
    ...props,
  }

  return mount(
    <MemoryRouter initialEntries={["/dont-match-anything"]} initialIndex={0}>
      <App {...mergedProps} />
    </MemoryRouter>
  )
}

describe("App", () => {
  it("renders 1 <Topbar> component and 2 <NavLink> components", () => {
    const wrapper = setup()
    expect(wrapper.find("Topbar").length).toBe(1)
    expect(wrapper.find("NavLink").length).toBe(2)
    // Render links to /tasks and /runs
    expect(
      wrapper
        .find("NavLink")
        .at(0)
        .props().to
    ).toBe("/tasks")
    expect(
      wrapper
        .find("NavLink")
        .at(1)
        .props().to
    ).toBe("/runs")
  })
  it("renders a modal if visible", () => {
    const fakeModal = "Some string will suffice."
    const wrapper = setup({
      modal: {
        modalVisible: true,
        modal: fakeModal,
      },
    })

    expect(wrapper.find("ModalContainer").length).toBe(1)
    expect(wrapper.find("ModalContainer").text()).toEqual(fakeModal)
  })
  it("renders a popup if visible", () => {
    const fakePopup = "Some string will suffice."
    const wrapper = setup({
      popup: {
        popupVisible: true,
        popup: fakePopup,
      },
    })

    expect(wrapper.find("App").text()).toContain(fakePopup)
  })
})
