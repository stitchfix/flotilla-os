import React from "react"
import { mount } from "enzyme"
import { mountToJson } from "enzyme-to-json"
import { TaskRuns } from "../TaskRuns"
import { RequestStatus } from "../Request"

describe("TaskRuns", () => {
  it("renders", () => {
    const wrapper = mount(
      <TaskRuns
        requestStatus={RequestStatus.NOT_READY}
        data={null}
        isLoading={false}
        error={false}
        updateSort={jest.fn()}
        updatePage={jest.fn()}
        updateFilter={jest.fn()}
      />
    )
    expect(mountToJson(wrapper)).toMatchSnapshot()
  })
})
