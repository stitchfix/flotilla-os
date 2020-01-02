import React from "react"
import { mount } from "enzyme"
import Creatable from "react-select/lib/Creatable"
import Connected, { GroupNameSelect } from "../GroupNameSelect"
import api from "../../api"

jest.mock("../../helpers/FlotillaClient")

describe("GroupNameSelect", () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it("renders a Select component", () => {
    const props = {
      options: [
        { label: "a", value: "a" },
        { label: "b", value: "b" },
        { label: "c", value: "c" },
      ],
      value: "a",
      onChange: jest.fn(),
    }
    const wrapper = mount(<GroupNameSelect {...props} isDisabled={false} />)
    const select = wrapper.find(Creatable)

    // Ensure <Select> component is rendered.
    expect(select).toHaveLength(1)

    // Ensure <Select> component has correct `options` prop.
    expect(select.prop("options")).toEqual(props.options)

    // Ensure <Select> component has correct `value` prop.
    expect(select.prop("value")).toEqual({
      label: props.value,
      value: props.value,
    })

    // Ensure props.onChange is called when <Select>'s onChange prop is
    // called.
    expect(props.onChange).toHaveBeenCalledTimes(0)
    const onChangeProp = select.prop("onChange")
    if (onChangeProp) {
      onChangeProp({ label: "b", value: "b" }, { action: "select-option" })
    }
    expect(props.onChange).toHaveBeenCalledTimes(1)
  })

  it("calls api.listGroups", () => {
    expect(api.listGroups).toHaveBeenCalledTimes(0)
    mount(<Connected value="" onChange={jest.fn()} isDisabled={false} />)
    expect(api.listGroups).toHaveBeenCalledTimes(1)
  })
})
