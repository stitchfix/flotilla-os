import * as React from "react"
import { shallow, mount, ReactWrapper } from "enzyme"
import { MemoryRouter as Router } from "react-router-dom"
import { DebounceInput } from "react-debounce-input"
import Navigation from "../Navigation/Navigation"
import { Tasks } from "./Tasks"
import StyledField from "../styled/Field"
import { flotillaUIRequestStates, flotillaUIIntents } from "../../types"
import { SortOrders } from "../ListRequest/ListRequest"
import DataTable from "../DataTable/DataTable"
import ReactSelectWrapper from "../ReactSelectWrapper/ReactSelectWrapper"
import formConfiguration from "../../helpers/formConfiguration"

const DEFAULT_PROPS = {
  queryParams: {},
  requestState: flotillaUIRequestStates.READY,
  error: false,
  inFlight: false,
  data: {},
  updatePage: (page: number) => {},
  updateSort: (key: string) => {},
  updateSearch: (key: string, value: any) => {},
  currentSortKey: "",
  currentSortOrder: SortOrders.ASC,
  currentPage: 1,
}

describe("Tasks", () => {
  describe("render", () => {
    let wrapper: ReactWrapper
    let updateSearch = jest.fn()
    beforeAll(() => {
      const routerWrapper = mount(
        <Router>
          <Tasks {...DEFAULT_PROPS} updateSearch={updateSearch} />
        </Router>
      )
      wrapper = routerWrapper.find(Tasks)
    })
    it("renders a Navigation component with a create task action", () => {
      expect(wrapper.find(Navigation).length).toEqual(1)
      expect(wrapper.find(Navigation).prop("actions").length).toEqual(1)
      expect(wrapper.find(Navigation).prop("actions")).toEqual([
        {
          isLink: true,
          href: "/tasks/create",
          text: "Create New Task",
          buttonProps: {
            intent: flotillaUIIntents.PRIMARY,
          },
        },
      ])
    })
    it("renders the correct filters", () => {
      const filters = wrapper.find(StyledField)

      // Ensure that 3 filters are rendered.
      expect(filters.length).toBe(3)

      // Get the filters
      const alias = filters.at(0)
      const group = filters.at(1)
      const image = filters.at(2)
      const aliasInput = alias.find(DebounceInput)
      const groupInput = group.find(ReactSelectWrapper)
      const imageInput = image.find(DebounceInput)

      // Ensure all the filters exist
      expect(aliasInput.length).toEqual(1)
      expect(groupInput.length).toEqual(1)
      expect(imageInput.length).toEqual(1)

      // Update all the filters to ensure that `props.updateSearch` is called
      expect(updateSearch).toHaveBeenCalledTimes(0)

      const aliasValue = "aliasV"
      aliasInput.prop("onChange")({ target: { value: aliasValue } })
      expect(updateSearch).toHaveBeenCalledTimes(1)
      expect(updateSearch).toHaveBeenCalledWith(
        formConfiguration.alias.key,
        aliasValue
      )

      const groupValue = "groupV"
      groupInput.prop("onChange")(groupValue)
      expect(updateSearch).toHaveBeenCalledTimes(2)
      expect(updateSearch).toHaveBeenCalledWith(
        formConfiguration.group_name.key,
        groupValue
      )

      const imageValue = "imageV"
      imageInput.prop("onChange")({ target: { value: imageValue } })
      expect(updateSearch).toHaveBeenCalledTimes(3)
      expect(updateSearch).toHaveBeenCalledWith(
        formConfiguration.image.key,
        imageValue
      )
    })
    it("renders a DataTable with the correct props", () => {
      expect(wrapper.find(DataTable).length).toBe(1)
    })
  })
})
