import React from "react"
import { mount, ReactWrapper } from "enzyme"
import { mountToJson } from "enzyme-to-json"
import { Formik } from "formik"
import { UpdateTaskForm } from "../UpdateTaskForm"
import { RequestStatus } from "../Request"

describe("UpdateTaskForm", () => {
  const onSubmit = jest.fn()
  const initialValues = {
    env: [],
    image: "",
    group_name: "",
    memory: 1024,
    command: "",
    tags: [],
  }
  let wrapper: ReactWrapper

  beforeEach(() => {
    wrapper = mount(
      <Formik initialValues={initialValues} onSubmit={onSubmit}>
        {({ values, setFieldValue, isValid }) => (
          <UpdateTaskForm
            values={values}
            isValid={isValid}
            setFieldValue={setFieldValue}
            requestStatus={RequestStatus.NOT_READY}
            error={null}
            isLoading={false}
            errors={{}}
          />
        )}
      </Formik>
    )
  })

  it("renders", () => {
    expect(mountToJson(wrapper)).toMatchSnapshot()
  })
})
