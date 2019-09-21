import * as React from "react"
import { mount, ReactWrapper } from "enzyme"
import { Button, ButtonGroup } from "@blueprintjs/core"
import Pagination, { Props } from "../Pagination"

describe("Pagination", () => {
  let wrapper: ReactWrapper<Props>

  beforeEach(() => {
    wrapper = mount(
      <Pagination
        updatePage={() => {}}
        currentPage={1}
        numItems={100}
        pageSize={20}
        isLoading={false}
      />
    )
  })

  it("renders two buttons", () => {
    expect(wrapper.find(ButtonGroup)).toHaveLength(1)
    expect(wrapper.find(Button)).toHaveLength(2)
  })

  it("disables the previous button if on the first page", () => {
    wrapper.setProps({ currentPage: 1 })
    expect(
      wrapper
        .find(Button)
        .at(0)
        .props().disabled
    ).toEqual(true)
  })

  it("disables the next button if on the last page", () => {
    wrapper.setProps({ numItems: 113, currentPage: 5 })
    expect(
      wrapper
        .find(Button)
        .at(1)
        .props().disabled
    ).toEqual(false)

    wrapper.setProps({ numItems: 113, currentPage: 6 })
    expect(
      wrapper
        .find(Button)
        .at(1)
        .props().disabled
    ).toEqual(true)
  })

  it("calls props.updatePage when the prev or next buttons are clicked", () => {
    const updatePage = jest.fn()

    wrapper.setProps({ updatePage, currentPage: 1 })
    expect(updatePage).toHaveBeenCalledTimes(0)
    wrapper
      .find(Button)
      .at(1)
      .simulate("click")
    expect(updatePage).toHaveBeenCalledTimes(1)
    expect(updatePage).toHaveBeenCalledWith(wrapper.props().currentPage + 1)

    wrapper.setProps({ currentPage: 2 })
    wrapper
      .find(Button)
      .at(0)
      .simulate("click")
    expect(updatePage).toHaveBeenCalledTimes(2)
    expect(updatePage).toHaveBeenCalledWith(wrapper.props().currentPage - 1)
  })
})
