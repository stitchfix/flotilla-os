import React from "react"
import { mount, ReactWrapper } from "enzyme"
import { Link, MemoryRouter } from "react-router-dom"
import ConnectedTasks, { Tasks, initialQuery } from "../Tasks"
import Request, { RequestStatus } from "../Request"
import { SortOrder } from "../../types"
import ViewHeader from "../ViewHeader"
import { DebounceInput } from "react-debounce-input"
import GroupNameSelect from "../GroupNameSelect"
import ListFiltersDropdown from "../ListFiltersDropdown"
import { Button } from "@blueprintjs/core"
import Pagination from "../Pagination"
import Table from "../Table"
import api from "../../api"

describe("Tasks", () => {
  it("renders a Request component and provides api.listTasks as the requestFn", () => {
    const wrapper = mount(
      <MemoryRouter>
        <ConnectedTasks />
      </MemoryRouter>
    )
    const r = wrapper.find(Request)
    expect(r).toHaveLength(1)
    expect(r.prop("requestFn")).toEqual(api.listTasks)
  })

  describe("rendering", () => {
    const aliasFilterValue = "my_alias"
    const groupNameFilterValue = "my_group"
    const imageFilterValue = "my_image"
    const updateSort = jest.fn()
    const updatePage = jest.fn()
    const updateFilter = jest.fn()
    let wrapper: ReactWrapper
    beforeAll(() => {
      wrapper = mount(
        <MemoryRouter>
          <Tasks
            requestStatus={RequestStatus.NOT_READY}
            data={null}
            isLoading={false}
            error={null}
            updateSort={updateSort}
            updatePage={updatePage}
            updateFilter={updateFilter}
            currentPage={1}
            currentSortKey="alias"
            currentSortOrder={SortOrder.ASC}
            query={{
              ...initialQuery,
              alias: aliasFilterValue,
              group_name: groupNameFilterValue,
              image: imageFilterValue,
            }}
          />
        </MemoryRouter>
      )
    })
    afterEach(() => {
      updateSort.mockReset()
      updatePage.mockReset()
      updateFilter.mockReset()
    })
    it("renders a ViewHeader component", () => {
      const vh = wrapper.find(ViewHeader)
      expect(vh).toHaveLength(1)
      expect(vh.prop("breadcrumbs")).toEqual([
        { text: "Tasks", href: "/tasks" },
      ])
      const createBtn = vh.find(Link)
      expect(createBtn).toHaveLength(1)
      expect(createBtn.prop("to")).toEqual(`/tasks/create`)
    })
    it("renders an alias filter", () => {
      const filter = wrapper.find(DebounceInput).filter("#tasksAliasFilter")
      expect(filter).toHaveLength(1)
      expect(filter.prop("value")).toEqual(aliasFilterValue)
      filter.prop("onChange")({ target: { value: "bar" } })
      expect(updateFilter).toHaveBeenCalledTimes(1)
      expect(updateFilter).toHaveBeenCalledWith("alias", "bar")
    })
    it("renders an group name filter", () => {
      wrapper
        .find(ListFiltersDropdown)
        .find(Button)
        .simulate("click")
      const filter = wrapper.find(GroupNameSelect)
      expect(filter).toHaveLength(1)
      expect(filter.prop("value")).toEqual(groupNameFilterValue)
      filter.prop("onChange")("bar")
      expect(updateFilter).toHaveBeenCalledTimes(1)
      expect(updateFilter).toHaveBeenCalledWith("group_name", "bar")
    })
    it("renders an image filter", () => {
      wrapper
        .find(ListFiltersDropdown)
        .find(Button)
        .simulate("click")
      const filter = wrapper.find(DebounceInput).filter("#tasksImageFilter")
      expect(filter).toHaveLength(1)
      expect(filter.prop("value")).toEqual(imageFilterValue)
      filter.prop("onChange")({ target: { value: "bar" } })
      expect(updateFilter).toHaveBeenCalledTimes(1)
      expect(updateFilter).toHaveBeenCalledWith("image", "bar")
    })
    it("renders pagination buttons", () => {
      expect(wrapper.find(Pagination)).toHaveLength(1)
    })
    it("renders a table", () => {
      const tb = wrapper.find(Table)
      expect(tb).toHaveLength(1)
      expect(tb.prop("columns")).toHaveProperty("alias")
      expect(tb.prop("columns")).toHaveProperty("group_name")
      expect(tb.prop("columns")).toHaveProperty("image")
      expect(tb.prop("columns")).toHaveProperty("memory")
    })
  })
})
