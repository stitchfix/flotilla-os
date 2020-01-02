import * as React from "react"
import { mount, ReactWrapper } from "enzyme"
import { ListRequest, Props, ChildProps } from "../ListRequest"
import { RequestStatus } from "../Request"
import { SortOrder } from "../../types"

const DEFAULT_PROPS: Props<any, any> = {
  requestStatus: RequestStatus.NOT_READY,
  data: null,
  isLoading: false,
  error: null,
  query: {},
  request: (args: any) => {},
  setQuery: (query: object, shouldReplace?: boolean) => {},
  initialQuery: {},
  getRequestArgs: (query: object) => {},
  children: (props: ChildProps<any, any>) => <span />,
  receivedAt: new Date(),
}

describe("ListRequest", () => {
  it("calls props.setQuery w/ props.initialQuery if props.query is empty on componentDidMount", () => {
    const realReq = ListRequest.prototype.request
    ListRequest.prototype.request = jest.fn()
    const setQuery = jest.fn()
    const initialQuery = { foo: "bar" }

    expect(setQuery).toHaveBeenCalledTimes(0)

    mount(
      <ListRequest
        {...DEFAULT_PROPS}
        initialQuery={initialQuery}
        query={{}}
        setQuery={setQuery}
      >
        {() => <span />}
      </ListRequest>
    )

    expect(setQuery).toHaveBeenCalledTimes(1)
    expect(setQuery).toHaveBeenCalledWith(initialQuery, true)
    expect(ListRequest.prototype.request).toHaveBeenCalledTimes(0)
    ListRequest.prototype.request = realReq
  })

  it("calls this.request if props.query is not empty on componentDidMount", () => {
    const realReq = ListRequest.prototype.request
    ListRequest.prototype.request = jest.fn()
    const setQuery = jest.fn()

    expect(setQuery).toHaveBeenCalledTimes(0)
    expect(ListRequest.prototype.request).toHaveBeenCalledTimes(0)

    const wrapper = mount(
      <ListRequest
        {...DEFAULT_PROPS}
        query={{ foo: "bar" }}
        setQuery={setQuery}
      >
        {() => <span />}
      </ListRequest>
    )

    expect(setQuery).toHaveBeenCalledTimes(0)
    expect(ListRequest.prototype.request).toHaveBeenCalledTimes(1)
    ListRequest.prototype.request = realReq
  })

  it("calls this.request if prevProps.query and props.query are not equal on componentDidUpdate", () => {
    const realReq = ListRequest.prototype.request
    ListRequest.prototype.request = jest.fn()
    expect(ListRequest.prototype.request).toHaveBeenCalledTimes(0)

    const wrapper = mount(
      <ListRequest {...DEFAULT_PROPS} query={{ foo: "bar" }}>
        {() => <span />}
      </ListRequest>
    )

    // Should have been called once when the component mounts.
    expect(ListRequest.prototype.request).toHaveBeenCalledTimes(1)

    wrapper.setProps({ query: { foo: "not-bar" } })

    expect(ListRequest.prototype.request).toHaveBeenCalledTimes(2)

    ListRequest.prototype.request = realReq
  })

  it("calls props.request with the correct args", () => {
    const request = jest.fn()
    const getRequestArgs = jest.fn(q => q)
    const query = { foo: "bar" }

    const wrapper = mount<ListRequest<any, any>>(
      <ListRequest
        {...DEFAULT_PROPS}
        request={request}
        getRequestArgs={getRequestArgs}
        query={query}
      >
        {() => <span />}
      </ListRequest>
    )

    const inst = wrapper.instance()

    expect(request).toHaveBeenCalledTimes(1)

    inst.request()
    expect(request).toHaveBeenCalledTimes(2)
    expect(request).toHaveBeenCalledWith(getRequestArgs(query))
  })

  it("calls props.children with the correct args", () => {
    const realUpdateSort = ListRequest.prototype.updateSort
    const realUpdatePage = ListRequest.prototype.updatePage
    const realUpdateFilter = ListRequest.prototype.updateFilter
    ListRequest.prototype.updateSort = jest.fn()
    ListRequest.prototype.updatePage = jest.fn()
    ListRequest.prototype.updateFilter = jest.fn()

    const wrapper = mount<ListRequest<any, any>>(
      <ListRequest {...DEFAULT_PROPS}>
        {(props: ChildProps<any, any>) => (
          <span>
            <button
              id="filter-btn"
              onClick={() => {
                props.updateFilter("foo", "bar")
              }}
            />
            <button
              id="page-btn"
              onClick={() => {
                props.updatePage(10)
              }}
            />
            <button
              id="sort-btn"
              onClick={() => {
                props.updateSort("a")
              }}
            />
          </span>
        )}
      </ListRequest>
    )

    // Test sort
    expect(ListRequest.prototype.updateSort).toHaveBeenCalledTimes(0)
    const sortButton = wrapper.find("#sort-btn")
    sortButton.simulate("click")
    expect(ListRequest.prototype.updateSort).toHaveBeenCalledTimes(1)
    expect(ListRequest.prototype.updateSort).toHaveBeenCalledWith("a")

    // Test page
    expect(ListRequest.prototype.updateFilter).toHaveBeenCalledTimes(0)
    const filterButton = wrapper.find("#filter-btn")
    filterButton.simulate("click")
    expect(ListRequest.prototype.updateFilter).toHaveBeenCalledTimes(1)
    expect(ListRequest.prototype.updateFilter).toHaveBeenCalledWith(
      "foo",
      "bar"
    )

    // Test filter
    expect(ListRequest.prototype.updatePage).toHaveBeenCalledTimes(0)
    const pageButton = wrapper.find("#page-btn")
    pageButton.simulate("click")
    expect(ListRequest.prototype.updatePage).toHaveBeenCalledTimes(1)
    expect(ListRequest.prototype.updatePage).toHaveBeenCalledWith(10)

    ListRequest.prototype.updateSort = realUpdateSort
    ListRequest.prototype.updatePage = realUpdatePage
    ListRequest.prototype.updateFilter = realUpdateFilter
  })

  describe("query update methods", () => {
    const setQuery = jest.fn()
    let wrapper: ReactWrapper<any>
    let instance: any

    beforeEach(() => {
      wrapper = mount<ListRequest<any, any>>(
        <ListRequest {...DEFAULT_PROPS} setQuery={setQuery} query={{ a: 1 }}>
          {() => <span />}
        </ListRequest>
      )
      instance = wrapper.instance() as ListRequest<any, any>
    })

    afterEach(() => {
      setQuery.mockReset()
    })

    it("updateSort calls setQuery with the correct arguments", () => {
      // Note: we're manually setting the wrapper's query prop since we're
      // mocking setQuery and it won't actually update the query.
      expect(setQuery).toHaveBeenCalledTimes(0)
      instance.updateSort("x")
      expect(setQuery).toHaveBeenCalledTimes(1)
      expect(setQuery).toHaveBeenCalledWith({
        ...wrapper.prop("query"),
        page: 1,
        sort_by: "x",
        order: SortOrder.ASC,
      })
      wrapper.setProps({ query: { sort_by: "x", order: SortOrder.ASC } })

      instance.updateSort("x")
      expect(setQuery).toHaveBeenCalledTimes(2)
      expect(setQuery).toHaveBeenCalledWith({
        ...wrapper.prop("query"),
        page: 1,
        sort_by: "x",
        order: SortOrder.DESC,
      })
      wrapper.setProps({ query: { sort_by: "x", order: SortOrder.DESC } })

      instance.updateSort("x")
      expect(setQuery).toHaveBeenCalledTimes(3)
      expect(setQuery).toHaveBeenCalledWith({
        ...wrapper.prop("query"),
        page: 1,
        sort_by: "x",
        order: SortOrder.ASC,
      })
      wrapper.setProps({ query: { sort_by: "x", order: SortOrder.ASC } })

      instance.updateSort("y")
      expect(setQuery).toHaveBeenCalledTimes(4)
      expect(setQuery).toHaveBeenCalledWith({
        ...wrapper.prop("query"),
        page: 1,
        sort_by: "y",
        order: SortOrder.ASC,
      })
    })

    it("updatePage calls setQuery with the correct arguments", () => {
      expect(setQuery).toHaveBeenCalledTimes(0)
      instance.updatePage(5000)
      expect(setQuery).toHaveBeenCalledTimes(1)
      expect(setQuery).toHaveBeenCalledWith({
        ...wrapper.prop("query"),
        page: 5000,
      })
    })

    it("updateFilter calls setQuery with the correct arguments", () => {
      expect(setQuery).toHaveBeenCalledTimes(0)
      instance.updateFilter("foo", "bar")
      expect(setQuery).toHaveBeenCalledTimes(1)
      expect(setQuery).toHaveBeenCalledWith({
        ...wrapper.prop("query"),
        page: 1,
        foo: "bar",
      })
    })
  })
})
