import React from "react"
import { mount, ReactWrapper } from "enzyme"
import { mountToJson } from "enzyme-to-json"
import Table from "../Table"

describe("Table", () => {
  type MockItem = {
    id: number
    name: string
  }
  let wrapper: ReactWrapper
  const updateSort = jest.fn()

  beforeEach(() => {
    wrapper = mount(
      <Table<MockItem>
        items={[
          { id: 1, name: "One" },
          { id: 2, name: "Two" },
          { id: 3, name: "Three" },
        ]}
        getItemKey={(i: MockItem) => i.id}
        updateSort={updateSort}
        columns={{
          id: {
            displayName: "ID",
            render: (item: MockItem) => item.id,
            isSortable: true,
          },
          name: {
            displayName: "Name",
            render: (item: MockItem) => item.name,
            isSortable: true,
          },
        }}
      />
    )
  })

  it("renders", () => {
    expect(mountToJson(wrapper)).toMatchSnapshot()
  })

  it("calls props.updateSort when a sortable column header is clicked", () => {})
})
