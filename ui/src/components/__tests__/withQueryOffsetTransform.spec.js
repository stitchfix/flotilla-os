import React from "react"
import { mount } from "enzyme"
import withQueryOffsetTransform from "../withQueryOffsetTransform"

describe("withQueryOffsetTransform", () => {
  it("logs an error if the options object is invalid", () => {
    const consoleErr = console.error
    console.error = jest.fn()

    expect(withQueryOffsetTransform()()).toBe(false)
    expect(console.error).toHaveBeenCalledTimes(1)
    console.error = consoleErr
  })
  it("intercepts any `updateQuery` calls and transforms the query before calling it's own props.updateQuery", () => {
    const Dummy = props => (
      <button
        onClick={() => {
          props.updateQuery({
            key: "offset",
            value: 40,
            updateType: "SHALLOW",
            replace: false,
          })
        }}
      />
    )
    const limit = 20
    const WrappedDummy = withQueryOffsetTransform(limit)(Dummy)
    const updateQuery = jest.fn()
    const wrapper = mount(
      <WrappedDummy
        updateQuery={updateQuery}
        query={{ page: 1, sort_by: "started_at", order: "desc" }}
        search="?page=1&sort_by=started_at&order=desc"
      />
    )
    const middleware = wrapper.find("Wrapped")
    const btn = wrapper.find("button")
    btn.simulate("click")
    expect(updateQuery).toHaveBeenCalledWith({
      key: "page",
      value: 3,
      updateType: "SHALLOW",
      replace: false,
    })
  })
  it("parses the query correctly", () => {
    const Dummy = () => <span />
    const limit = 20
    const WrappedDummy = withQueryOffsetTransform(limit)(Dummy)
    const query = { page: 2, sort_by: "started_at", order: "desc" }
    const search = "?page=1&sort_by=started_at&order=desc"
    const props = {
      updateQuery: () => {},
      query,
      search,
    }
    const wrapper = mount(<WrappedDummy {...props} />)
    expect(wrapper.find("Dummy").props().query).toEqual({
      limit,
      offset: 20,
      sort_by: "started_at",
      order: "desc",
    })
  })
})
