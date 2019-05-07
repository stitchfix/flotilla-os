import * as React from "react"
import { mount, ReactWrapper } from "enzyme"
import { ListRequest, SortOrders, IProps, IChildProps } from "./ListRequest"
import { flotillaUIRequestStates } from "../../types"

const DEFAULT_PROPS: IProps = {
  getRequestArgs: (query: any) => {},
  initialQuery: {},
  limit: 50,
  requestFn: (arg: any) =>
    new Promise(resolve => {
      resolve()
    }),
  shouldContinuouslyFetch: false,
  children: (props: IChildProps) => <span />,
  data: [],
  inFlight: false,
  requestState: flotillaUIRequestStates.NOT_READY,
  request: (args?: any) => {},
  error: false,
  queryParams: {},
  setQueryParams: (query: object, shouldReplace: boolean) => {},
}

describe("ListRequest", () => {
  describe("Lifecycle Methods", () => {
    describe("componentDidMount", () => {
      it("calls props.setQueryParams with props.initialQuery to trigger a request call on the next tick if props.queryParams is empty", () => {
        const q = {}
        const setQueryParams = jest.fn()
        mount(
          <ListRequest
            {...DEFAULT_PROPS}
            queryParams={q}
            setQueryParams={setQueryParams}
          />
        )
        expect(setQueryParams).toHaveBeenCalledTimes(1)
        expect(setQueryParams).toHaveBeenCalledWith(q, true)
      })
      it("calls this.requestData if props.queryParams is not empty", () => {
        const _requestData = ListRequest.prototype.requestData
        ListRequest.prototype.requestData = jest.fn()

        mount(<ListRequest {...DEFAULT_PROPS} queryParams={{ foo: "bar" }} />)
        expect(ListRequest.prototype.requestData).toHaveBeenCalledTimes(1)
        ListRequest.prototype.requestData = _requestData
      })
      it("calls window.setInterval if props.shouldContinuouslyFetch is true", () => {
        const wrapper = mount(
          <ListRequest {...DEFAULT_PROPS} shouldContinuouslyFetch />
        )
        const inst = wrapper.instance() as ListRequest
        expect(inst.requestInterval).not.toBeNaN()
      })
    })

    describe("componentDidUpdate", () => {
      it("calls this.requestData if the props.queryParams has changed", () => {
        const _requestData = ListRequest.prototype.requestData
        ListRequest.prototype.requestData = jest.fn()
        const wrapper = mount(
          <ListRequest {...DEFAULT_PROPS} queryParams={{ a: 1 }} />
        )
        expect(ListRequest.prototype.requestData).toHaveBeenCalledTimes(1)

        wrapper.setProps({ queryParams: { a: 2 } })
        expect(ListRequest.prototype.requestData).toHaveBeenCalledTimes(2)
        ListRequest.prototype.requestData = _requestData
      })
    })
    describe("componentWillUnmount", () => {
      it("calls this.clearInterval", () => {
        const _clearInterval = ListRequest.prototype.clearInterval
        ListRequest.prototype.clearInterval = jest.fn()
        const wrapper = mount(
          <ListRequest {...DEFAULT_PROPS} queryParams={{ a: 1 }} />
        )
        expect(ListRequest.prototype.clearInterval).toHaveBeenCalledTimes(0)

        wrapper.unmount()
        expect(ListRequest.prototype.clearInterval).toHaveBeenCalledTimes(1)
        ListRequest.prototype.clearInterval = _clearInterval
      })
    })
  })

  describe("Query Updating Methods", () => {
    let setQueryParams: jest.Mock
    let wrapper: ReactWrapper
    let inst: ListRequest

    beforeEach(() => {
      setQueryParams = jest.fn()
      wrapper = mount(
        <ListRequest
          {...DEFAULT_PROPS}
          queryParams={{ a: 1 }}
          setQueryParams={setQueryParams}
        />
      )
      inst = wrapper.instance() as ListRequest
    })

    describe("updatePage", () => {
      it("calls this.props.setQueryParams with the correct parameters", () => {
        inst.updatePage(5)
        expect(setQueryParams).toHaveBeenCalledTimes(1)
        expect(setQueryParams).toHaveBeenCalledWith({ page: 5 }, false)
      })
      it("does not call this.props.setQuery if the `page` arg is less than 1", () => {
        inst.updatePage(-1)
        expect(setQueryParams).toHaveBeenCalledTimes(0)
      })
    })

    describe("updateSearch", () => {
      it("calls this.props.setQueryParams with the correct parameters", () => {
        const k = "foo"
        const v = "bar"
        inst.updateSearch(k, v)
        expect(setQueryParams).toHaveBeenCalledTimes(1)
        expect(setQueryParams).toHaveBeenCalledWith({ [k]: v }, false)
      })
    })

    describe("updateSort", () => {
      it("calls this.props.setQueryParams with the correct parameters", () => {
        // Handle case where we need to reverse the sort order for the same key
        let sortKey = "alias"
        let sortOrder = SortOrders.ASC
        wrapper = mount(
          <ListRequest
            {...DEFAULT_PROPS}
            queryParams={{
              sort_by: sortKey,
              order: sortOrder,
            }}
            setQueryParams={setQueryParams}
          />
        )
        inst = wrapper.instance() as ListRequest

        // Update once with the same sortKey, the order should now be
        // reversed.
        inst.updateSort(sortKey)
        let next = {
          sort_by: sortKey,
          order: SortOrders.DESC,
        }
        expect(setQueryParams).toHaveBeenCalledTimes(1)
        expect(setQueryParams).toHaveBeenCalledWith(next, false)
        // Manually set query since `setQueryParams` is a mock
        wrapper.setProps({ queryParams: next })

        // Update again with the same sortKey and reverse again.
        inst.updateSort(sortKey)
        next = {
          sort_by: sortKey,
          order: SortOrders.ASC,
        }
        expect(setQueryParams).toHaveBeenCalledTimes(2)
        expect(setQueryParams).toHaveBeenCalledWith(next, false)
        wrapper.setProps({ queryParams: next })

        // Update with a new sortKey, order should be asc.
        const nextSortKey = "baz"
        inst.updateSort(nextSortKey)
        next = {
          sort_by: nextSortKey,
          order: SortOrders.ASC,
        }
        expect(setQueryParams).toHaveBeenCalledTimes(3)
        expect(setQueryParams).toHaveBeenCalledWith(next, false)
      })
    })
  })

  describe("requestData", () => {
    it("calls props.request with the correct argument", () => {
      const getRequestArgs = jest.fn(query => query)
      const queryParams = { page: 1 }
      const request = jest.fn()
      const limit = 10
      const wrapper = mount(
        <ListRequest
          {...DEFAULT_PROPS}
          limit={limit}
          getRequestArgs={getRequestArgs}
          queryParams={queryParams}
          request={request}
        />
      )
      const inst = wrapper.instance() as ListRequest

      // It should be called once when the component mounts.
      expect(request).toHaveBeenCalledTimes(1)

      // Manually call requestData.
      inst.requestData()

      expect(request).toHaveBeenCalledTimes(2)
      expect(request).toHaveBeenCalledWith(
        getRequestArgs(ListRequest.preprocessQuery(queryParams, limit))
      )
    })
  })

  describe("Static Methods", () => {
    describe("preprocessQuery", () => {
      it("replaces the `page` attribute with `offset` and `limit` attributes", () => {
        expect(
          ListRequest.preprocessQuery({ page: 5, foo: "bar" }, 10)
        ).toEqual({
          foo: "bar",
          offset: 40,
          limit: 10,
        })
      })
    })
  })
})
