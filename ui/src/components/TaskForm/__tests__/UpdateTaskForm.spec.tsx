import * as React from "react"
import { mount } from "enzyme"
import { MemoryRouter } from "react-router-dom"
import { get } from "lodash"
import { UpdateTaskForm, WithTaskContext, IProps } from "../UpdateTaskForm"
import api from "../../../api"
import {
  flotillaUIRequestStates,
  IFlotillaUITaskContext,
  IFlotillaTaskDefinition,
  IFlotillaAPIError,
} from "../../../types"
import TaskContext from "../../Task/TaskContext"

describe("UpdateTaskForm", () => {
  describe("BaseUpdateTaskForm", () => {
    const mockDefinitionID = "definitionID"
    const submitValues = {
      command: "command",
      env: [],
      group_name: "group_name",
      image: "image",
      memory: 1024,
      tags: [],
    }

    const defaultProps: IProps = {
      push: () => {},
      renderPopup: () => {},
      defaultValues: submitValues,
      definitionID: mockDefinitionID,
      title: "",
    }
    const realUpdateTask = api.updateTask

    beforeAll(() => {})
    beforeEach(() => {
      api.updateTask = jest.fn()
    })

    afterEach(() => {
      api.updateTask = realUpdateTask
    })

    it("renders a BaseTaskFormWithSelectOptions component", async () => {
      const wrapper = mount(
        <MemoryRouter>
          <UpdateTaskForm {...defaultProps} />
        </MemoryRouter>
      )
      expect(wrapper.find("BaseTaskFormWithSelectOptions").length).toBe(1)
    })

    it("calls api.updateTask when the form is submitted", () => {
      const wrapper = mount(
        <MemoryRouter>
          <UpdateTaskForm {...defaultProps} />
        </MemoryRouter>
      )
      const inst = wrapper.find("UpdateTaskForm").instance() as UpdateTaskForm
      expect(api.updateTask).toHaveBeenCalledTimes(0)
      inst.handleSubmit(submitValues)
      expect(api.updateTask).toHaveBeenCalledTimes(1)
      expect(api.updateTask).toHaveBeenCalledWith({
        definitionID: mockDefinitionID,
        values: submitValues,
      })
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
          <UpdateTaskForm {...defaultProps} push={push} />
        </MemoryRouter>
      )
      expect(push).toHaveBeenCalledTimes(0)
      const inst = wrapper.find("UpdateTaskForm").instance() as UpdateTaskForm
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
          <UpdateTaskForm {...defaultProps} renderPopup={renderPopup} />
        </MemoryRouter>
      )
      expect(renderPopup).toHaveBeenCalledTimes(0)
      const inst = wrapper.find("UpdateTaskForm").instance() as UpdateTaskForm
      inst.handleFail(error)
      expect(renderPopup).toHaveBeenCalledTimes(1)
    })
  })

  describe("WithTaskContext", () => {
    const defaultCtx: IFlotillaUITaskContext = {
      data: null,
      inFlight: false,
      error: false,
      requestState: flotillaUIRequestStates.NOT_READY,
      definitionID: "",
      requestData: () => {},
    }

    const defaultProps = {
      renderPopup: () => {},
      push: () => {},
    }
    it("renders a loader if task data hasn't been fetched", () => {
      const wrapper = mount(
        <MemoryRouter>
          <TaskContext.Provider value={defaultCtx}>
            <WithTaskContext {...defaultProps} />
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
            <WithTaskContext {...defaultProps} />
          </TaskContext.Provider>
        </MemoryRouter>
      )

      let baseTaskFormWrapper = wrapper.find("BaseTaskFormWithSelectOptions")
      expect(baseTaskFormWrapper.length).toBe(1)
      expect(baseTaskFormWrapper.prop("defaultValues")).toEqual({
        command: get(ctx, ["data", "command"], ""),
        env: get(ctx, ["data", "env"], []),
        group_name: get(ctx, ["data", "group_name"], ""),
        image: get(ctx, ["data", "image"], ""),
        memory: get(ctx, ["data", "memory"], 1024),
        tags: get(ctx, ["data", "tags"], []),
      })
      expect(baseTaskFormWrapper.prop("title")).toBe(
        `Copy Task ${get(ctx, ["data", "alias"], ctx.definitionID)}`
      )
    })
  })
})
