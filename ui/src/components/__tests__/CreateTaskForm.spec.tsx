import React from "react"
import { mount, ReactWrapper } from "enzyme"
import { Formik, FastField } from "formik"
import { AxiosError } from "axios"
import { Button, Callout } from "@blueprintjs/core"
import { CreateTaskForm } from "../CreateTaskForm"
import BaseTaskForm from "../BaseTaskForm"
import { RequestStatus } from "../Request"

describe("CreateTaskForm", () => {
  const onSubmit = jest.fn()
  const initialValues = {
    env: [],
    image: "",
    group_name: "",
    memory: 1024,
    command: "",
    tags: [],
    alias: "",
  }
  let wrapper: ReactWrapper

  beforeEach(() => {
    wrapper = mount(
      <Formik initialValues={initialValues} onSubmit={onSubmit}>
        {({ values, setFieldValue, isValid, errors }) => (
          <CreateTaskForm
            values={values}
            isValid={isValid}
            setFieldValue={setFieldValue}
            requestStatus={RequestStatus.NOT_READY}
            error={null}
            isLoading={false}
            errors={errors}
          />
        )}
      </Formik>
    )
  })

  it("renders an error callout if there's an error", () => {
    const e: AxiosError = {
      config: {},
      name: "error",
      message: "error_message",
    }
    const errWrapper = mount(
      <Formik initialValues={initialValues} onSubmit={onSubmit}>
        {({ values, setFieldValue, isValid, errors }) => (
          <CreateTaskForm
            values={values}
            isValid={isValid}
            setFieldValue={setFieldValue}
            requestStatus={RequestStatus.ERROR}
            error={e}
            isLoading={false}
            errors={errors}
          />
        )}
      </Formik>
    )
    expect(errWrapper.find(Callout).length).toBe(1)
  })

  it("renders an `alias` field", () => {
    const fields = wrapper.find(FastField)
    expect(fields.at(0).prop("name")).toEqual("alias")
  })

  it("renders a BaseTaskForm component", () => {
    expect(wrapper.find(BaseTaskForm).length).toBe(1)
  })

  it("renders a submit button", () => {
    const btns = wrapper.find(Button)
    expect(btns.length).toBeGreaterThanOrEqual(1)
    const submitBtn = btns.at(btns.length - 1)
    expect(submitBtn)
  })
})
