import * as React from "react"
import { mount } from "enzyme"
import { get } from "lodash"
import { MemoryRouter } from "react-router-dom"
import CopyTaskForm from "../CopyTaskForm"
import TaskContext from "../../Task/TaskContext"
import { flotillaUIRequestStates, IFlotillaUITaskContext } from "../../../types"

const defaultCtx: IFlotillaUITaskContext = {
  data: null,
  inFlight: false,
  error: false,
  requestState: flotillaUIRequestStates.NOT_READY,
  definitionID: "",
  requestData: () => {},
}

describe("CopyTaskForm", () => {
  it("renders a loader if task data hasn't been fetched", () => {
    const wrapper = mount(
      <MemoryRouter>
        <TaskContext.Provider value={defaultCtx}>
          <CopyTaskForm />
        </TaskContext.Provider>
      </MemoryRouter>
    )
    expect(wrapper.find("Loader").length).toBe(1)
    expect(wrapper.find("CreateTaskForm").length).toBe(0)
  })

  it("renders a CreateTaskForm with the correct props if task data has been fetched", () => {
    const ctx: IFlotillaUITaskContext = {
      ...defaultCtx,
      data: {
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
      },
      requestState: flotillaUIRequestStates.READY,
    }
    const wrapper = mount(
      <MemoryRouter>
        <TaskContext.Provider value={ctx}>
          <CopyTaskForm />
        </TaskContext.Provider>
      </MemoryRouter>
    )

    let createTaskFormWrapper = wrapper.find("CreateTaskForm")
    expect(createTaskFormWrapper.length).toBe(1)
    expect(createTaskFormWrapper.prop("defaultValues")).toEqual({
      alias: get(ctx, ["data", "alias"], ""),
      command: get(ctx, ["data", "command"], ""),
      env: get(ctx, ["data", "env"], []),
      group_name: get(ctx, ["data", "group_name"], ""),
      image: get(ctx, ["data", "image"], ""),
      memory: get(ctx, ["data", "memory"], 1024),
      tags: get(ctx, ["data", "tags"], []),
    })
    expect(createTaskFormWrapper.prop("title")).toBe(
      `Copy Task ${get(ctx, ["data", "alias"], ctx.definitionID)}`
    )
  })
})
