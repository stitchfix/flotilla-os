import React from "react"
import { configureSetup } from "../../__testutils__"
import FlotillaTopbar from "../FlotillaTopbar"

const setup = configureSetup({
  unconnected: FlotillaTopbar,
})

describe("FlotillaTopbar", () => {
  let wrapper
  beforeAll(() => {
    wrapper = setup({ connectToRouter: true })
  })
  it("renders the logo and app name", () => {
    expect(wrapper.find("Link.pl-topbar-app-name").length).toBe(1)
    expect(wrapper.find("Link.pl-topbar-app-name").props().to).toBe("/")
    expect(wrapper.find("img.topbar-logo").length).toBe(1)
    expect(wrapper.find("Link.pl-topbar-app-name").text()).toBe("FLOTILLA")
  })
  it("renders two NavLink components", () => {
    expect(wrapper.find("NavLink").length).toBe(2)
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
})
