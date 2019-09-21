import * as React from "react"
import { mount } from "enzyme"
import { Formik, FastField } from "formik"
import { FormGroup } from "@blueprintjs/core"
import {
  groupNameFieldSpec,
  imageFieldSpec,
  commandFieldSpec,
  memoryFieldSpec,
  tagsFieldSpec,
  envFieldSpec,
} from "../../constants"
import BaseTaskForm from "../BaseTaskForm"
import EnvFieldArray from "../EnvFieldArray"
import { Env } from "../../types"
import FieldError from "../FieldError"

jest.mock("../../helpers/FlotillaClient")

describe("BaseTaskForm", () => {
  it("renders the correct fields", () => {
    const groupNameInitialValue = "my_group_name"
    const imageInitialValue = "my_image"
    const commandInitialValue = "my_command"
    const memoryInitialValue = 1024
    const tagsInitialValue = ["a", "b", "c"]
    const envInitialValue: Env[] = []
    const wrapper = mount(
      <Formik
        initialValues={{
          [groupNameFieldSpec.name]: groupNameInitialValue,
          [imageFieldSpec.name]: imageInitialValue,
          [commandFieldSpec.name]: commandInitialValue,
          [memoryFieldSpec.name]: memoryInitialValue,
          [tagsFieldSpec.name]: tagsInitialValue,
          [envFieldSpec.name]: envInitialValue,
        }}
        onSubmit={jest.fn()}
      >
        {({ values, setFieldValue, errors }) => {
          return (
            <BaseTaskForm
              values={values}
              setFieldValue={setFieldValue}
              errors={errors}
            />
          )
        }}
      </Formik>
    )

    const formGroups = wrapper.find(FormGroup)
    const fields = wrapper.find(FastField)

    // Ensure that components have the correct lengths.
    expect(formGroups).toHaveLength(5)
    expect(fields).toHaveLength(5)
    expect(wrapper.find(EnvFieldArray)).toHaveLength(1)
    expect(wrapper.find(FieldError)).toHaveLength(0)

    // Group name field.
    expect(formGroups.at(0).props().label).toEqual(groupNameFieldSpec.label)
    expect(formGroups.at(0).props().helperText).toEqual(
      groupNameFieldSpec.description
    )
    expect(fields.at(0).props().name).toEqual(groupNameFieldSpec.name)
    expect(fields.at(0).props().value).toEqual(groupNameInitialValue)

    // Image field.
    expect(formGroups.at(1).props().label).toEqual(imageFieldSpec.label)
    expect(formGroups.at(1).props().helperText).toEqual(
      imageFieldSpec.description
    )
    expect(fields.at(1).props().name).toEqual(imageFieldSpec.name)
    expect(
      fields
        .at(1)
        .find("input")
        .props().value
    ).toEqual(imageInitialValue)

    // Command field.
    expect(formGroups.at(2).props().label).toEqual(commandFieldSpec.label)
    expect(formGroups.at(2).props().helperText).toEqual(
      commandFieldSpec.description
    )
    expect(fields.at(2).props().name).toEqual(commandFieldSpec.name)
    expect(
      fields
        .at(2)
        .find("textarea")
        .props().value
    ).toEqual(commandInitialValue)

    // Memory field.
    expect(formGroups.at(3).props().label).toEqual(memoryFieldSpec.label)
    expect(formGroups.at(3).props().helperText).toEqual(
      memoryFieldSpec.description
    )
    expect(fields.at(3).props().name).toEqual(memoryFieldSpec.name)
    expect(
      fields
        .at(3)
        .find("input")
        .props().value
    ).toEqual(memoryInitialValue)

    // Tags field.
    expect(formGroups.at(4).props().label).toEqual(tagsFieldSpec.label)
    expect(formGroups.at(4).props().helperText).toEqual(
      tagsFieldSpec.description
    )
    expect(fields.at(4).props().name).toEqual(tagsFieldSpec.name)
    expect(fields.at(4).props().value).toEqual(tagsInitialValue)
  })
})
