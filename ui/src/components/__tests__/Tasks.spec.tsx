import React from "react"
import { mount } from "enzyme"
import { MemoryRouter } from "react-router-dom"
import ConnectedTasks, {
  Tasks as UnconnectedTasks,
  Props,
  initialQuery,
} from "../Tasks"
import { RequestStatus } from "../Request"
import ListRequest from "../ListRequest"
import { SortOrder } from "../../types"
import { Spinner } from "@blueprintjs/core"
import Table from "../Table"
import api from "../../api"
import ErrorCallout from "../ErrorCallout"
import { createMockTaskObject } from "../../helpers/testHelpers"

jest.mock("../../helpers/FlotillaClient")

describe("Tasks", () => {
  describe("Connected", () => {
    it("renders ListRequest and provides api.listTasks as the requestFn", () => {
      expect(api.listTasks).toHaveBeenCalledTimes(0)

      const wrapper = mount(
        <MemoryRouter>
          <ConnectedTasks />
        </MemoryRouter>
      )

      expect(wrapper.find(ListRequest)).toHaveLength(1)
      expect(wrapper.find(ListRequest).prop("requestFn")).toEqual(api.listTasks)
      expect(api.listTasks).toHaveBeenCalledTimes(1)
    })
  })

  describe("Unconnected", () => {
    const defaultProps: Props = {
      requestStatus: RequestStatus.NOT_READY,
      data: null,
      isLoading: false,
      error: null,
      updateSort: () => {},
      updatePage: () => {},
      updateFilter: () => {},
      currentPage: 1,
      currentSortKey: "alias",
      currentSortOrder: SortOrder.ASC,
      query: initialQuery,
    }

    it("renders a Spinner props.requestStatus is `NOT_READY`", () => {
      const wrapper = mount(
        <MemoryRouter>
          <UnconnectedTasks
            {...defaultProps}
            requestStatus={RequestStatus.NOT_READY}
          />
        </MemoryRouter>
      )
      expect(wrapper.find(ErrorCallout)).toHaveLength(0)
      expect(wrapper.find(Table)).toHaveLength(0)
      expect(wrapper.find(Spinner)).toHaveLength(1)
    })

    it("renders a Table props.requestStatus is `READY`", () => {
      const wrapper = mount(
        <MemoryRouter>
          <UnconnectedTasks
            {...defaultProps}
            requestStatus={RequestStatus.READY}
            data={{
              offset: 0,
              limit: 20,
              total: 20,
              definitions: [
                createMockTaskObject({ definition_id: "a" }),
                createMockTaskObject({ definition_id: "b" }),
                createMockTaskObject({ definition_id: "c" }),
              ],
            }}
          />
        </MemoryRouter>
      )
      expect(wrapper.find(ErrorCallout)).toHaveLength(0)
      expect(wrapper.find(Spinner)).toHaveLength(0)
      expect(wrapper.find(Table)).toHaveLength(1)
      expect(wrapper.find(Table).prop("columns")).toHaveProperty("alias")
      expect(wrapper.find(Table).prop("columns")).toHaveProperty("group_name")
      expect(wrapper.find(Table).prop("columns")).toHaveProperty("image")
      expect(wrapper.find(Table).prop("columns")).toHaveProperty("memory")
    })

    it("renders an ErrorCallout props.requestStatus is `ERROR`", () => {
      const wrapper = mount(
        <MemoryRouter>
          <UnconnectedTasks
            {...defaultProps}
            requestStatus={RequestStatus.ERROR}
          />
        </MemoryRouter>
      )
      expect(wrapper.find(ErrorCallout)).toHaveLength(1)
      expect(wrapper.find(Table)).toHaveLength(0)
      expect(wrapper.find(Spinner)).toHaveLength(0)
    })
  })
})
