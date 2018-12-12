import * as React from "react"
import { mount, EnzymeAdapter, ReactWrapper } from "enzyme"
import { Formik, FormikProps, Field } from "formik"
import FormikKVField from "../FormikKVField"
import { intents } from "../../../.."

describe("FormikKVField", () => {
  let wrapper: ReactWrapper<any>
  const keyField = "name"
  const valueField = "value"
  const initialValue = [
    { [keyField]: "a", [valueField]: "1" },
    { [keyField]: "b", [valueField]: "2" },
    { [keyField]: "c", [valueField]: "3" },
  ]
  beforeAll(() => {
    wrapper = mount(
      <Formik
        initialValues={{ env: initialValue }}
        onSubmit={(values, actions) => {}}
        render={(f: FormikProps<{ env: any[] }>) => (
          <form onSubmit={f.handleSubmit}>
            <FormikKVField
              name="env"
              value={f.values.env}
              keyField={keyField}
              valueField={valueField}
              setFieldValue={f.setFieldValue}
              label="Test"
              description="test"
              isKeyRequired
              isValueRequired={false}
            />
          </form>
        )}
      />
    )
  })
  it("renders a KVFieldInput component with the correct props", () => {
    expect(wrapper.find("KVFieldInput")).toHaveLength(1)
  })
  it("renders the existing values and remove button", () => {
    const rows = wrapper.find("NestedKeyValueRow")
    expect(rows).toHaveLength(initialValue.length + 1)

    for (let i = 0; i < initialValue.length; i++) {
      const fields = rows.at(i).find(Field)
      const removeButton = rows.at(i).find("Button")
      expect(fields).toHaveLength(2)
      expect(fields.at(0).prop("value")).toEqual(initialValue[i][keyField])
      expect(fields.at(1).prop("value")).toEqual(initialValue[i][valueField])
      expect(removeButton).toHaveLength(1)
      expect(removeButton.prop("type")).toEqual("button")
      expect(removeButton.prop("intent")).toEqual(intents.ERROR)
    }
  })
})
