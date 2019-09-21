import * as React from "react"
import flushPromiseQueue from "flush-promises"
import { mount, ReactWrapper } from "enzyme"
import CreateTaskForm, {
  ConnectedProps as Props,
  CreateTaskForm as UnconnectedCreateTaskForm,
} from "../CreateTaskForm"
import api from "../../api"
import { Formik } from "formik"
import {
  createMockRouteComponentProps,
  mockFormikActions,
} from "../../helpers/testHelpers"
import Request from "../Request"
import BaseTaskForm from "../BaseTaskForm"

jest.mock("../../helpers/FlotillaClient")

describe("CreateTaskForm", () => {
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
    initialValues: {
      env: [{ name: "foo", value: "bar" }],
      image: "my_image",
      group_name: "my_group",
      alias: "my_alias",
      memory: 1024,
      command: "my_command",
      tags: ["a", "b"],
    },
    onSuccess: jest.fn(),
  }

  let wrapper: ReactWrapper

  beforeEach(() => {
    jest.clearAllMocks()
    wrapper = mount(<CreateTaskForm {...props} />)
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
    ).toBe(api.createTask)
    expect(
      wrapper
        .find(Request)
        .at(0)
        .props().shouldRequestOnMount
    ).toEqual(false)

    expect(wrapper.find(Formik)).toHaveLength(1)
    expect(wrapper.find(UnconnectedCreateTaskForm)).toHaveLength(1)
    expect(wrapper.find(BaseTaskForm)).toHaveLength(1)
    expect(wrapper.find('input[name="alias"]')).toHaveLength(1)
    expect(wrapper.find("button#submitButton")).toHaveLength(1)
  })

  it("calls api.createTask when submitted", async () => {
    // At this point, we don't expect any functions to have been called.
    expect(api.createTask).toHaveBeenCalledTimes(0)
    expect(props.onSuccess).toHaveBeenCalledTimes(0)
    expect(props.history.push).toHaveBeenCalledTimes(0)

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
    expect(api.createTask).toHaveBeenCalledTimes(1)

    // Flush the promise queue.
    await flushPromiseQueue()

    // Expect `onSuccess` and `push` to be invoked once.
    expect(props.onSuccess).toHaveBeenCalledTimes(1)
    expect(props.history.push).toHaveBeenCalledTimes(1)
  })
})
