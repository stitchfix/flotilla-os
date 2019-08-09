import React from "react"
import { mount } from "enzyme"
import { mountToJson } from "enzyme-to-json"
import { GroupNameSelect } from "../GroupNameSelect"

describe("GroupNameSelect", () => {
  it("renders", () => {
    const wrapper = mount(
      <GroupNameSelect value="" onChange={jest.fn()} options={[]} />
    )
    expect(mountToJson(wrapper)).toMatchSnapshot()
  })
})
