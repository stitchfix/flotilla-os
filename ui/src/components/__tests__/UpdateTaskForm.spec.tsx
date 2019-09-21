import * as React from "react"
import flushPromiseQueue from "flush-promises"
import { mount, ReactWrapper } from "enzyme"
import UpdateTaskForm, {
  ConnectedProps as Props,
  UpdateTaskForm as UnconnectedUpdateTaskForm,
} from "../UpdateTaskForm"
import api from "../../api"
import { Formik } from "formik"
import {
  createMockRouteComponentProps,
  mockFormikActions,
  createMockTaskObject,
} from "../../helpers/testHelpers"
import Request, { RequestStatus } from "../Request"
import BaseTaskForm from "../BaseTaskForm"
import { TaskContext, TaskCtx as TaskContextTypeDef } from "../Task"

jest.mock("../../helpers/FlotillaClient")

describe("UpdateTaskForm", () => {
  const DEFINITION_ID = "my_def_id"

  // Instantiate mock route component props object.
  const mockRouteComponentProps = createMockRouteComponentProps({
    path: "/tasks/create",
    url: "/tasks/create",
    params: {},
  })

  // Instantiate props object.
  const props: Props = {
    ...mockRouteComponentProps,
    history: {
      ...mockRouteComponentProps.history,
      push: jest.fn(),
    },
    definitionID: DEFINITION_ID,
  }

  // Instantiate context object.
  const mockTaskCtx: TaskContextTypeDef = {
    data: createMockTaskObject({ definition_id: DEFINITION_ID }),
    requestStatus: RequestStatus.READY,
    isLoading: false,
    error: null,
    request: jest.fn(),
    basePath: "",
    definitionID: DEFINITION_ID,
  }

  let wrapper: ReactWrapper

  beforeEach(() => {
    jest.clearAllMocks()
    wrapper = mount(
      <TaskContext.Provider value={mockTaskCtx}>
        <UpdateTaskForm {...props} />
      </TaskContext.Provider>
    )
  })

  it("renders the correct components", () => {
    // Note: there will be more than 1 Request component due to those wrapping
    // GroupNameSelect, etc.
    expect(wrapper.find(Request).length).toBeGreaterThanOrEqual(1)
    expect(
      wrapper
        .find(Request)
        .at(0)
        .props().requestFn
    ).toBe(api.updateTask)
    expect(
      wrapper
        .find(Request)
        .at(0)
        .props().shouldRequestOnMount
    ).toEqual(false)

    expect(wrapper.find(Formik)).toHaveLength(1)
    expect(wrapper.find(UnconnectedUpdateTaskForm)).toHaveLength(1)
    expect(wrapper.find(BaseTaskForm)).toHaveLength(1)
    expect(wrapper.find("button#submitButton")).toHaveLength(1)
  })

  it("calls api.updateTask when submitted", async () => {
    // At this point, we don't expect any functions to have been called.
    expect(api.updateTask).toHaveBeenCalledTimes(0)
    expect(props.history.push).toHaveBeenCalledTimes(0)
    expect(mockTaskCtx.request).toHaveBeenCalledTimes(0)

    // Manually invoke Formik's onSubmit prop.
    wrapper
      .find(Formik)
      .props()
      .onSubmit(
        {
          env: [{ name: "foo", value: "bar" }],
          image: "my_image",
          group_name: "my_group",
          alias: "my_alias",
          memory: 1024,
          command: "my_command",
          tags: ["a", "b"],
        },
        mockFormikActions
      )

    // Expect FlotillaClient's `createTask` method to be invoked once.
    expect(api.updateTask).toHaveBeenCalledTimes(1)

    // Flush the promise queue.
    await flushPromiseQueue()

    // Expect `onSuccess` and `push` to be invoked once.
    expect(props.history.push).toHaveBeenCalledTimes(1)
    expect(mockTaskCtx.request).toHaveBeenCalledTimes(1)
  })
})
