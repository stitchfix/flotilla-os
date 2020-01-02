import * as React from "react"
import { mount, shallow } from "enzyme"
import { LogVirtualized, Props } from "../LogVirtualized"

const defaultProps: Props = {
  width: 100,
  height: 100,
  logs: ["a", "b", "c", "d"],
  shouldAutoscroll: true,
  dispatch: jest.fn(),
}

describe("LogVirtualized", () => {
  it("scrolls to the most recent line upon mounting", () => {
    const scrollTo = LogVirtualized.prototype.scrollTo
    LogVirtualized.prototype.scrollTo = jest.fn()

    expect(LogVirtualized.prototype.scrollTo).toHaveBeenCalledTimes(0)

    // Mount LogVirtualized with shouldAutoscroll === true.
    shallow(<LogVirtualized {...defaultProps} />)

    expect(LogVirtualized.prototype.scrollTo).toHaveBeenCalledTimes(1)

    // Mount LogVirtualized with shouldAutoscroll === false.
    shallow(<LogVirtualized {...defaultProps} shouldAutoscroll={false} />)

    expect(LogVirtualized.prototype.scrollTo).toHaveBeenCalledTimes(1)
    LogVirtualized.prototype.scrollTo = scrollTo
  })

  it("calls this.handleCursorChange if state.searchCursor is updated", () => {
    const handleCursorChange = LogVirtualized.prototype.handleCursorChange
    LogVirtualized.prototype.handleCursorChange = jest.fn()
    expect(LogVirtualized.prototype.handleCursorChange).toHaveBeenCalledTimes(0)
    const wrapper = mount(<LogVirtualized {...defaultProps} />)
    wrapper.setState({ searchCursor: 10 })
    expect(LogVirtualized.prototype.handleCursorChange).toHaveBeenCalledTimes(1)
    LogVirtualized.prototype.handleCursorChange = handleCursorChange
  })

  it("scrolls to the most recent line if the number of lines is different", () => {
    const scrollTo = LogVirtualized.prototype.scrollTo
    LogVirtualized.prototype.scrollTo = jest.fn()
    const wrapper = mount(<LogVirtualized {...defaultProps} />)
    expect(LogVirtualized.prototype.scrollTo).toHaveBeenCalledTimes(1)
    wrapper.setProps({ logs: ["a", "b", "c", "d", "e", "f"] })
    expect(LogVirtualized.prototype.scrollTo).toHaveBeenCalledTimes(2)
    LogVirtualized.prototype.scrollTo = scrollTo
  })

  it("handles search correctly", () => {
    const logs = ["one two three", "four five six", "seven eight nine"]
    const wrapper = mount<LogVirtualized>(
      <LogVirtualized {...defaultProps} logs={logs} />
    )
    expect(wrapper.state().searchMatches).toEqual([])
    expect(wrapper.state().searchCursor).toEqual(0)
    let query = "s"
    wrapper.instance().search(query)
    expect(wrapper.state().searchMatches).toEqual([
      [1, logs[1].indexOf(query)],
      [2, logs[2].indexOf(query)],
    ])
    expect(wrapper.state().searchCursor).toEqual(0)

    query = "seven"
    wrapper.instance().search(query)
    expect(wrapper.state().searchMatches).toEqual([[2, logs[2].indexOf(query)]])
    expect(wrapper.state().searchCursor).toEqual(0)
  })

  it("handles cursor changes correctly", () => {
    const scrollTo = LogVirtualized.prototype.scrollTo
    LogVirtualized.prototype.scrollTo = jest.fn()
    const fn = LogVirtualized.prototype.scrollTo as jest.Mock
    const wrapper = mount<LogVirtualized>(<LogVirtualized {...defaultProps} />)
    const searchMatches: [number, number][] = [
      [0, 0],
      [1, 0],
      [2, 0],
      [3, 0],
    ]

    wrapper.setState({ searchMatches })

    let cursor = 1
    wrapper.setState({ searchCursor: cursor })
    expect(fn.mock.calls[fn.mock.calls.length - 1]).toEqual([
      searchMatches[cursor][0],
      "center",
    ])

    cursor = 2
    wrapper.setState({ searchCursor: cursor })
    expect(fn.mock.calls[fn.mock.calls.length - 1]).toEqual([
      searchMatches[cursor][0],
      "center",
    ])

    LogVirtualized.prototype.scrollTo = scrollTo
  })
})
