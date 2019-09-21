import * as React from "react"
import { mount, ReactWrapper } from "enzyme"
import { Formik, FastField } from "formik"
import { Button } from "@blueprintjs/core"
import { EnvFieldArray } from "../EnvFieldArray"
import { Env } from "../../types"

describe("EnvFieldArray", () => {
  let wrapper: ReactWrapper
  const values: Env[] = [
    { name: "a", value: "b" },
    { name: "c", value: "d" },
    { name: "e", value: "f" },
  ]
  const push = jest.fn()
  const remove = jest.fn()

  beforeAll(() => {
    wrapper = mount(
      <Formik initialValues={{ env: values }} onSubmit={jest.fn()}>
        {() => <EnvFieldArray values={values} push={push} remove={remove} />}
      </Formik>
    )
  })

  it("renders props.values", () => {
    const items = wrapper.find(".flotilla-env-field-array-item")
    expect(items).toHaveLength(values.length)
    for (let i = 0; i < items.length; i++) {
      const item: ReactWrapper = items.at(i)
      expect(item.find(FastField)).toHaveLength(2)
      expect(item.find("button")).toHaveLength(1)
    }
  })

  it("calls props.remove with the index of the item when clicked", () => {
    // Get the second item
    const index = 1
    const second = wrapper.find(".flotilla-env-field-array-item").at(index)
    expect(remove).toHaveBeenCalledTimes(0)
    second.find("button").simulate("click")
    expect(remove).toHaveBeenCalledTimes(1)
    expect(remove).toHaveBeenCalledWith(index)
  })

  it("calls props.push with an empty env struct when the add button is clicked", () => {
    const addButton = wrapper
      .find(Button)
      .filterWhere(r => r.hasClass("flotilla-env-field-array-add-button"))
    expect(push).toHaveBeenCalledTimes(0)
    addButton.simulate("click")
    expect(push).toHaveBeenCalledTimes(1)
    expect(push).toHaveBeenCalledWith({ name: "", value: "" })
  })
})
