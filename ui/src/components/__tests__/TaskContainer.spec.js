import React from "react"
import { shallow } from "enzyme"
import fetchMock from "fetch-mock"
import WrappedTaskContainer, { TaskContainer } from "../TaskContainer"

const definitionId = "asdf"
const res = { foo: "bar" }
fetchMock.get("*", res)

describe("TaskContainer", () => {
  let realFetch = TaskContainer.prototype.fetch
  beforeEach(() => {
    TaskContainer.prototype.fetch = jest.fn()
  })
  afterEach(() => {
    TaskContainer.prototype.fetch = realFetch
  })
  it("calls props.fetch with the router's `definitionId` param when the component mounts", () => {
    const wrapper = shallow(
      <TaskContainer match={{ params: { definitionId } }} />
    )
    expect(TaskContainer.prototype.fetch).toHaveBeenCalledTimes(1)
    expect(TaskContainer.prototype.fetch).toHaveBeenCalledWith(
      expect.stringContaining(definitionId)
    )
  })
  it("calls props.fetch with nextProps.match.params.definitionId when the definitionId changes", () => {
    const wrapper = shallow(
      <TaskContainer match={{ params: { definitionId } }} />
    )
    expect(TaskContainer.prototype.fetch).toHaveBeenCalledTimes(1)

    const nextDefinitionId = "nextDefinitionId"
    wrapper.setProps({ match: { params: { definitionId: nextDefinitionId } } })
    expect(TaskContainer.prototype.fetch).toHaveBeenCalledTimes(2)
    expect(TaskContainer.prototype.fetch).toHaveBeenCalledWith(
      expect.stringContaining(nextDefinitionId)
    )
  })
})
