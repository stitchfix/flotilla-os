import React from "react"
import { mount } from "enzyme"
import { MemoryRouter } from "react-router-dom"
import ConnectedRuns, {
  Runs as UnconnectedRuns,
  Props,
  initialQuery,
} from "../Runs"
import { RequestStatus } from "../Request"
import ListRequest from "../ListRequest"
import { SortOrder } from "../../types"
import { Spinner } from "@blueprintjs/core"
import Table from "../Table"
import api from "../../api"
import ErrorCallout from "../ErrorCallout"
import { createMockRunObject } from "../../helpers/testHelpers"

jest.mock("../../helpers/FlotillaClient")

describe("Runs", () => {
  describe("Connected", () => {
    it("renders ListRequest and provides api.listRun as the requestFn", () => {
      expect(api.listRun).toHaveBeenCalledTimes(0)

      const wrapper = mount(
        <MemoryRouter>
          <ConnectedRuns />
        </MemoryRouter>
      )

      expect(wrapper.find(ListRequest)).toHaveLength(1)
      expect(wrapper.find(ListRequest).prop("requestFn")).toEqual(api.listRun)
      expect(api.listRun).toHaveBeenCalledTimes(1)
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
      currentSortKey: "started_at",
      currentSortOrder: SortOrder.DESC,
      query: initialQuery,
      receivedAt: new Date(),
    }

    it("renders a Spinner props.requestStatus is `NOT_READY`", () => {
      const wrapper = mount(
        <MemoryRouter>
          <UnconnectedRuns
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
          <UnconnectedRuns
            {...defaultProps}
            requestStatus={RequestStatus.READY}
            data={{
              offset: 0,
              limit: 20,
              total: 3,
              history: [
                createMockRunObject({ run_id: "a" }),
                createMockRunObject({ run_id: "b" }),
                createMockRunObject({ run_id: "c" }),
              ],
            }}
          />
        </MemoryRouter>
      )
      expect(wrapper.find(ErrorCallout)).toHaveLength(0)
      expect(wrapper.find(Spinner)).toHaveLength(0)
      expect(wrapper.find(Table)).toHaveLength(1)
      expect(wrapper.find(Table).prop("columns")).toHaveProperty("status")
      expect(wrapper.find(Table).prop("columns")).toHaveProperty("started_at")
      expect(wrapper.find(Table).prop("columns")).toHaveProperty("run_id")
      expect(wrapper.find(Table).prop("columns")).toHaveProperty("alias")
    })

    it("renders an ErrorCallout props.requestStatus is `ERROR`", () => {
      const wrapper = mount(
        <MemoryRouter>
          <UnconnectedRuns
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
