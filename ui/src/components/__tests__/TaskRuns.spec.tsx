import React from "react"
import { mount } from "enzyme"
import { MemoryRouter } from "react-router-dom"
import ConnectedTaskRuns, {
  TaskRuns as UnconnectedTaskRuns,
  Props,
  initialQuery,
} from "../TaskRuns"
import { RequestStatus } from "../Request"
import ListRequest from "../ListRequest"
import { SortOrder } from "../../types"
import { Spinner } from "@blueprintjs/core"
import Table from "../Table"
import api from "../../api"
import ErrorCallout from "../ErrorCallout"
import { createMockRunObject } from "../../helpers/testHelpers"

jest.mock("../../helpers/FlotillaClient")

describe("TaskRuns", () => {
  describe("Connected", () => {
    it("renders ListRequest and provides api.listTaskRuns as the requestFn", () => {
      const definitionID = "foo"
      expect(api.listTaskRuns).toHaveBeenCalledTimes(0)

      const wrapper = mount(
        <MemoryRouter>
          <ConnectedTaskRuns definitionID={definitionID} />
        </MemoryRouter>
      )

      expect(wrapper.find(ListRequest)).toHaveLength(1)
      expect(wrapper.find(ListRequest).prop("requestFn")).toEqual(
        api.listTaskRuns
      )
      expect(api.listTaskRuns).toHaveBeenCalledTimes(1)
      expect(api.listTaskRuns).toHaveBeenCalledWith(
        expect.objectContaining({
          definitionID,
        })
      )
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
      receivedAt: new Date(),
    }

    it("renders a Spinner props.requestStatus is `NOT_READY`", () => {
      const wrapper = mount(
        <MemoryRouter>
          <UnconnectedTaskRuns
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
          <UnconnectedTaskRuns
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
      expect(wrapper.find(Table).prop("columns")).toHaveProperty("run_id")
      expect(wrapper.find(Table).prop("columns")).toHaveProperty("status")
      expect(wrapper.find(Table).prop("columns")).toHaveProperty("started_at")
      expect(wrapper.find(Table).prop("columns")).toHaveProperty("finished_at")
      expect(wrapper.find(Table).prop("columns")).toHaveProperty("cluster")
    })

    it("renders an ErrorCallout props.requestStatus is `ERROR`", () => {
      const wrapper = mount(
        <MemoryRouter>
          <UnconnectedTaskRuns
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
