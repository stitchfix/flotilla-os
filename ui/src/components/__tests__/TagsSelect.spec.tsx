import React from "react"
import { mount } from "enzyme"
import { mountToJson } from "enzyme-to-json"
import { TagsSelect } from "../TagsSelect"

describe("TagsSelect", () => {
  it("renders", () => {
    const wrapper = mount(
      <TagsSelect value={[]} onChange={jest.fn()} options={[]} />
    )
    expect(mountToJson(wrapper)).toMatchSnapshot()
  })
})
