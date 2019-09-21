import * as React from "react"
import { mount } from "enzyme"
import { MemoryRouter } from "react-router-dom"
import qs from "qs"
import ConnectedQueryParams from "../QueryParams"

describe("QueryParams", () => {
  it("provides a `query` and `setQuery` prop to it's children", () => {
    const children = jest.fn(() => <span />)
    const q = "?foo=bar&bar=baz&env=a|b&env=c|d"
    const wrapper = mount(
      <MemoryRouter
        initialEntries={[
          {
            pathname: "foo",
            search: q,
          },
        ]}
      >
        <ConnectedQueryParams>{children}</ConnectedQueryParams>
      </MemoryRouter>
    )
    expect(children).toHaveBeenCalledWith({
      query: qs.parse(q.substr(1)),
      setQuery: expect.any(Function),
    })
  })
})
