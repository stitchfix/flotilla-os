import * as React from "react"
import { mount } from "enzyme"
import { MemoryRouter } from "react-router-dom"
import { CreateTaskForm } from "../CreateTaskForm"
import api from "../../../api"
import { IFlotillaTaskDefinition, IFlotillaAPIError } from "../../../types"

const submitValues = {
  alias: "alias",
  command: "command",
  env: [],
  group_name: "group_name",
  image: "image",
  memory: 1024,
  tags: [],
}

describe("CreateTaskForm", () => {
  const realCreateTask = api.createTask

  beforeAll(() => {})
  beforeEach(() => {
    api.createTask = jest.fn()
  })

  afterEach(() => {
    api.createTask = realCreateTask
  })

  it("renders a BaseTaskFormWithSelectOptions component", async () => {
    const wrapper = mount(
      <MemoryRouter>
        <CreateTaskForm />
      </MemoryRouter>
    )
    expect(wrapper.find("BaseTaskFormWithSelectOptions").length).toBe(1)
  })

  it("calls api.createTask when the form is submitted", () => {
    const wrapper = mount(
      <MemoryRouter>
        <CreateTaskForm />
      </MemoryRouter>
    )
    const inst = wrapper.find("CreateTaskForm").instance() as CreateTaskForm
    expect(api.createTask).toHaveBeenCalledTimes(0)
    inst.handleSubmit(submitValues)
    expect(api.createTask).toHaveBeenCalledTimes(1)
    expect(api.createTask).toHaveBeenCalledWith({ values: submitValues })
  })

  it("redirects to the created task if created successfully", async () => {
    const taskDef: IFlotillaTaskDefinition = {
      alias: "alias",
      arn: "arn",
      command: "command",
      container_name: "container_name",
      definition_id: "definition_id",
      env: [],
      group_name: "group_name",
      image: "image",
      memory: 128,
      tags: [],
    }
    const push = jest.fn()
    const wrapper = mount(
      <MemoryRouter>
        <CreateTaskForm push={push} />
      </MemoryRouter>
    )
    expect(push).toHaveBeenCalledTimes(0)
    const inst = wrapper.find("CreateTaskForm").instance() as CreateTaskForm
    inst.handleSuccess(taskDef)
    expect(push).toHaveBeenCalledTimes(1)
    expect(push).toHaveBeenCalledWith(`/tasks/${taskDef.definition_id}`)
  })

  it("renders a popup if not created successfully", () => {
    const renderPopup = jest.fn()
    const error: IFlotillaAPIError = {
      data: "Failed",
    }
    const wrapper = mount(
      <MemoryRouter>
        <CreateTaskForm renderPopup={renderPopup} />
      </MemoryRouter>
    )
    expect(renderPopup).toHaveBeenCalledTimes(0)
    const inst = wrapper.find("CreateTaskForm").instance() as CreateTaskForm
    inst.handleFail(error)
    expect(renderPopup).toHaveBeenCalledTimes(1)
  })
})
