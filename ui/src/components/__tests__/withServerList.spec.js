import React from "react"
import { MemoryRouter } from "react-router-dom"
import { mount } from "enzyme"
import qs from "query-string"
import {
  withRouterSync,
  queryUpdateTypes,
  withStateFetch,
} from "aa-ui-components"
import withQueryOffsetTransform from "../withQueryOffsetTransform"
import withServerList from "../withServerList"

const baseProps = {}
const DummyComponent = props => {
  return <div>dummy</div>
}
DummyComponent.displayName = "DummyComponent"
const wrappedDisplayName = `withServerList(${DummyComponent.displayName})`

const wrapperSetup = opts => withServerList(opts)(DummyComponent)

const createWithServerListComponent = opts => wrapperSetup(opts).withoutHOCStack

const mountSetup = (component, props = {}, routerProps = {}) => {
  const WithHOCStack = withRouterSync(
    withQueryOffsetTransform(20)(withStateFetch(component))
  )

  return mount(
    <MemoryRouter {...routerProps}>
      <WithHOCStack {...props} />
    </MemoryRouter>
  )
}

describe("withServerList", () => {
  describe("Lifecycle Methods", () => {
    describe("componentDidMount", () => {
      const WithServerListComponent = createWithServerListComponent({
        getUrl: () => {},
        limit: 20,
        defaultQuery: { offset: 0, sort_by: "alias", order: "asc" },
      })
      const _fetch = WithServerListComponent.prototype.fetch
      const _cwrp = WithServerListComponent.prototype.componentWillReceiveProps
      beforeEach(() => {
        WithServerListComponent.prototype.fetch = jest.fn()
        WithServerListComponent.prototype.componentWillReceiveProps = jest.fn()
      })
      afterEach(() => {
        WithServerListComponent.prototype.fetch = _fetch
        WithServerListComponent.prototype.componentWillReceiveProps = _cwrp
      })
      it("calls this.fetch if props.query exists", () => {
        const query = { offset: "0" }
        mountSetup(
          WithServerListComponent,
          {},
          {
            initialEntries: [
              {
                pathname: "/tasks",
                search: `?${qs.stringify(query)}`,
              },
            ],
            initialIndex: 0,
          }
        )

        expect(WithServerListComponent.prototype.fetch).toHaveBeenCalledTimes(1)
      })
      it("calls props.updateQuery with props.defaultQuery and doesn't call this.fetch if props.query is empty", () => {
        // Mock componentWillReceiveProps to ensure that it doesn't call `fetch`
        mountSetup(WithServerListComponent)
        expect(WithServerListComponent.prototype.fetch).toHaveBeenCalledTimes(0)
      })
    })
    describe("componentWillReceiveProps", () => {
      const WithServerListComponent = createWithServerListComponent({
        getUrl: () => {},
        limit: 20,
        defaultQuery: { offset: 0, sort_by: "alias", order: "asc" },
      })
      const _fetch = WithServerListComponent.prototype.fetch
      beforeEach(() => {
        WithServerListComponent.prototype.fetch = jest.fn()
      })
      afterEach(() => {
        WithServerListComponent.prototype.fetch = _fetch
      })
      it("calls this.fetch is props.query has changed", () => {
        const query = { offset: "0" }
        const wrapper = mountSetup(
          WithServerListComponent,
          {},
          {
            initialEntries: [
              {
                pathname: "/tasks",
                search: `?${qs.stringify(query)}`,
              },
            ],
            initialIndex: 0,
          }
        ).find(wrappedDisplayName)
        // Call once when component mounts
        expect(WithServerListComponent.prototype.fetch).toHaveBeenCalledTimes(1)

        wrapper.props().updateQuery({
          key: "offset",
          value: "20",
          updateType: queryUpdateTypes.SHALLOW,
        })
        expect(WithServerListComponent.prototype.fetch).toHaveBeenCalledTimes(2)
      })
    })
  })
})
