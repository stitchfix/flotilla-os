import * as React from "react"
import { mount, ReactWrapper } from "enzyme"
import { DebounceInput } from "react-debounce-input"
import { ButtonGroup, Button } from "@blueprintjs/core"
import LogVirtualizedSearch from "../LogVirtualizedSearch"

describe("LogVirtualizedSearch", () => {
  let wrapper: ReactWrapper
  const onChange = jest.fn()
  const onFocus = jest.fn()
  const onBlur = jest.fn()
  const onIncrement = jest.fn()
  const onDecrement = jest.fn()
  beforeAll(() => {
    wrapper = mount(
      <LogVirtualizedSearch
        onChange={onChange}
        onFocus={onFocus}
        onBlur={onBlur}
        onIncrement={onIncrement}
        onDecrement={onDecrement}
        inputRef={null}
        cursorIndex={0}
        totalMatches={0}
      />
    )
  })
  it("renders the correct components", () => {
    expect(
      wrapper.find(".flotilla-logs-virtualized-search-container")
    ).toHaveLength(1)
    expect(wrapper.find(DebounceInput)).toHaveLength(1)
    expect(wrapper.find(Button)).toHaveLength(2)
  })

  it("handles input events", () => {
    const input = wrapper.find(DebounceInput)
    expect(onFocus).toHaveBeenCalledTimes(0)
    expect(onBlur).toHaveBeenCalledTimes(0)
    input.simulate("focus")
    expect(onFocus).toHaveBeenCalledTimes(1)
    expect(onBlur).toHaveBeenCalledTimes(0)
    input.simulate("blur")
    expect(onFocus).toHaveBeenCalledTimes(1)
    expect(onBlur).toHaveBeenCalledTimes(1)
  })

  it("handles button click events", () => {
    wrapper.setProps({ cursorIndex: 5, totalMatches: 20 })
    const buttons = wrapper.find(Button)
    expect(onIncrement).toHaveBeenCalledTimes(0)
    expect(onDecrement).toHaveBeenCalledTimes(0)
    buttons.at(0).simulate("click")
    expect(onIncrement).toHaveBeenCalledTimes(0)
    expect(onDecrement).toHaveBeenCalledTimes(1)
    buttons.at(1).simulate("click")
    expect(onIncrement).toHaveBeenCalledTimes(1)
    expect(onDecrement).toHaveBeenCalledTimes(1)
  })
})
