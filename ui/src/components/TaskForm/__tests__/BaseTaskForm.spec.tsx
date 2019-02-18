jest.mock("../../../helpers/FlotillaAPIClient")

import * as React from "react"
import { mount, ReactWrapper } from "enzyme"
import { MemoryRouter } from "react-router-dom"
import * as flushPromises from "flush-promises"
import { BaseTaskForm, IProps, TaskFormPayload } from "../BaseTaskForm"
import { CreateTaskYupSchema } from "../validation"

/** Default values used to initialize the form. */
const defaultValues: TaskFormPayload = {
  alias: "",
  memory: 1024,
  group_name: "",
  image: "",
  command: "",
  env: [],
  tags: [],
}

/** Default BaseTaskForm props for convenience. */
const defaultProps: IProps = {
  defaultValues,
  groupOptions: [],
  tagOptions: [],
  title: "Title",
  submitFn: (values: TaskFormPayload) =>
    new Promise(resolve => {
      resolve()
    }),
  onSuccess: (opts?: any) => {},
  onFail: (error?: any) => {},
  validateSchema: CreateTaskYupSchema,
  shouldRenderAliasField: true,
}

describe("BaseTaskForm", () => {
  describe("Rendering", () => {
    let wrapper: ReactWrapper
    beforeAll(() => {
      wrapper = mount(
        <MemoryRouter>
          <BaseTaskForm {...defaultProps} />
        </MemoryRouter>
      )
    })

    it("renders a View component", () => {
      expect(wrapper.find("View").length).toBe(1)
    })

    it("renders a Navigation component", () => {
      const navWrapper = wrapper.find("ConnectedTaskFormNavigation")
      const btnsWrapper = navWrapper.find("Button")
      expect(navWrapper.length).toBe(1)
      expect(btnsWrapper.length).toBe(2)
    })

    it("renders a Formik component", () => {
      expect(wrapper.find("Formik").length).toBe(1)
    })

    it("renders the correct field components", () => {
      expect(wrapper.find("AliasField").length).toBe(1)
      expect(wrapper.find("GroupNameField").length).toBe(1)
      expect(wrapper.find("ImageField").length).toBe(1)
      expect(wrapper.find("CommandField").length).toBe(1)
      expect(wrapper.find("MemoryField").length).toBe(1)
      expect(wrapper.find("TagsField").length).toBe(1)
      expect(wrapper.find("FormikKVField").length).toBe(1)
      expect(wrapper.find("FormikKVField").prop("name")).toBe("env")
    })
  })

  describe("Submitting", () => {
    const submitValues = {
      alias: "alias",
      command: "command",
      env: [],
      group_name: "group_name",
      image: "image",
      memory: 1024,
      tags: [],
    }

    it("calls its submitFn prop when submitted", () => {
      const submitFn = jest.fn(
        (values: TaskFormPayload) =>
          new Promise(resolve => {
            resolve()
          })
      )
      const wrapper = mount(
        <MemoryRouter>
          <BaseTaskForm {...defaultProps} submitFn={submitFn} />
        </MemoryRouter>
      )

      // Find the instance of BaseTaskForm.
      const formWrapperInst = wrapper
        .find(BaseTaskForm)
        .instance() as BaseTaskForm

      // submitFn shouldn't be called yet.
      expect(submitFn).toHaveBeenCalledTimes(0)

      // Manually call handleSubmit.
      formWrapperInst.handleSubmit(submitValues)

      // submitFn should have been called with the values.
      expect(submitFn).toHaveBeenCalledTimes(1)
      expect(submitFn).toHaveBeenCalledWith(submitValues)
    })
    it("can handle successful submissions", async () => {
      const response = { foo: "bar" }
      const submitFn = jest.fn(
        (values: TaskFormPayload) =>
          new Promise(resolve => {
            resolve(response)
          })
      )
      const onFail = jest.fn()
      const onSuccess = jest.fn()
      const wrapper = mount(
        <MemoryRouter>
          <BaseTaskForm
            {...defaultProps}
            submitFn={submitFn}
            onFail={onFail}
            onSuccess={onSuccess}
          />
        </MemoryRouter>
      )

      // Find the instance of BaseTaskForm.
      const formWrapperInst = wrapper
        .find(BaseTaskForm)
        .instance() as BaseTaskForm

      // Neither onSuccess nor onFail should have been called yet.
      expect(onSuccess).toHaveBeenCalledTimes(0)
      expect(onFail).toHaveBeenCalledTimes(0)
      expect(formWrapperInst.state.inFlight).toBe(false)
      expect(formWrapperInst.state.error).toBe(false)

      // Call handleSubmit.
      formWrapperInst.handleSubmit(submitValues)

      // Immediately after calling handleSubmit, state.inFlight should be set
      // to true.
      expect(formWrapperInst.state.inFlight).toBe(true)
      expect(formWrapperInst.state.error).toBe(false)

      // Wait for promises to resolve.
      await flushPromises()

      // state.inFlight should now be false and the onSuccess callback should
      // have been called with the server's response.
      expect(formWrapperInst.state.inFlight).toBe(false)
      expect(formWrapperInst.state.error).toBe(false)
      expect(onFail).toHaveBeenCalledTimes(0)
      expect(onSuccess).toHaveBeenCalledTimes(1)
      expect(onSuccess).toHaveBeenCalledWith(response)
    })
    it("can handle unsuccessful submissions", async () => {
      const error = "ERROR"
      const submitFn = jest.fn(
        () =>
          new Promise((resolve, reject) => {
            reject(error)
          })
      )
      const onFail = jest.fn()
      const onSuccess = jest.fn()
      const wrapper = mount(
        <MemoryRouter>
          <BaseTaskForm
            {...defaultProps}
            submitFn={submitFn}
            onFail={onFail}
            onSuccess={onSuccess}
          />
        </MemoryRouter>
      )

      // Find the instance of BaseTaskForm.
      const formWrapperInst = wrapper
        .find(BaseTaskForm)
        .instance() as BaseTaskForm

      // Neither onSuccess nor onFail should have been called yet.
      expect(onSuccess).toHaveBeenCalledTimes(0)
      expect(onFail).toHaveBeenCalledTimes(0)
      expect(formWrapperInst.state.inFlight).toBe(false)
      expect(formWrapperInst.state.error).toBe(false)

      // Call handleSubmit.
      formWrapperInst.handleSubmit(submitValues)

      // Immediately after calling handleSubmit, state.inFlight should be set
      // to true.
      expect(formWrapperInst.state.inFlight).toBe(true)
      expect(formWrapperInst.state.error).toBe(false)

      // Wait for promises to resolve.
      await flushPromises()

      // state.inFlight should now be false and the onFail callback should have
      // been called with the error returned by the server.
      expect(formWrapperInst.state.inFlight).toBe(false)
      expect(formWrapperInst.state.error).toBe(error)
      expect(onSuccess).toHaveBeenCalledTimes(0)
      expect(onFail).toHaveBeenCalledTimes(1)
      expect(onFail).toHaveBeenCalledWith(error)
    })
  })
})
