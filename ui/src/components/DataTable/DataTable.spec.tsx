import * as React from "react"
import { mount, ReactWrapper } from "enzyme"
import DataTable, { IProps } from "./DataTable"
import { TableRow } from "../styled/Table"
import { SortOrders } from "../ListRequest/ListRequest"
import { flotillaUIRequestStates } from "../../types"
import Loader from "../styled/Loader"

const DEFAULT_PROPS: IProps = {
  items: ["a", "b", "c"],
  columns: {
    columnA: {
      allowSort: false,
      displayName: "Column A",
      render: (item: any) => item,
      width: 1,
    },
  },
  getItemKey: (_: string, index: number) => index,
  onSortableHeaderClick: (sortKey: string) => {},
  currentSortKey: "",
  currentSortOrder: SortOrders.ASC,
  currentPage: 1,
}

describe("DataTable", () => {
  let wrapper: ReactWrapper
  beforeAll(() => {
    wrapper = mount(<DataTable {...DEFAULT_PROPS} />)
  })

  it("renders a Table with props.items as rows", () => {
    wrapper.setProps({ requestState: flotillaUIRequestStates.READY })
    expect(wrapper.find(Loader).length).toEqual(0)
    const tbody = wrapper.find("tbody")
    expect(tbody.length).toBe(1)
    const trow = tbody.find(TableRow)
    expect(trow.length).toBe(DEFAULT_PROPS.items.length)
  })
})
