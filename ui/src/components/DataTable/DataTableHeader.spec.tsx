import * as React from "react"
import { mount } from "enzyme"
import {
  TableHeaderCell,
  TableHeaderCellSortable,
  TableHeaderSortIcon,
} from "../styled/Table"
import DataTableHeader, { IProps } from "./DataTableHeader"
import { SortOrders } from "../ListRequest/ListRequest"

const DEFAULT_PROPS: IProps = {
  isSortable: false,
  currentSortKey: "key",
  currentSortOrder: SortOrders.ASC,
  children: "test",
  sortKey: "key",
  onClick: () => {},
  width: 1,
}

describe("DataTableColumnHeader", () => {
  it("renders a TableHeaderCell if props.isSortable is false", () => {
    const wrapper = mount(
      <DataTableHeader {...DEFAULT_PROPS} isSortable={false} />
    )
    expect(wrapper.find(TableHeaderCell).length).toEqual(1)
    expect(wrapper.find(TableHeaderCellSortable).length).toEqual(0)
  })
  it("renders a TableHeaderCellSortable if props.isSortable is true", () => {
    const wrapper = mount(<DataTableHeader {...DEFAULT_PROPS} isSortable />)
    expect(wrapper.find(TableHeaderCell).length).toEqual(0)
    expect(wrapper.find(TableHeaderCellSortable).length).toEqual(1)
  })
  it("renders the children prop", () => {
    const dn = "hi"
    const wrapper = mount(
      <DataTableHeader {...DEFAULT_PROPS}>{dn}</DataTableHeader>
    )
    expect(wrapper.prop("children")).toEqual(dn)
  })
  it("sets TableHeaderCellSortable's isActive prop to true is props.currentSortKey equals the column's key", () => {
    const KEY = "KEY"
    const wrapper = mount(
      <DataTableHeader
        {...DEFAULT_PROPS}
        isSortable
        currentSortKey={KEY}
        sortKey={KEY}
      />
    )
    expect(wrapper.find(TableHeaderCellSortable).prop("isActive")).toEqual(true)
  })
})
