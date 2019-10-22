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
  cpuFieldSpec,
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
    const cpuInitialValue = 512
    const tagsInitialValue = ["a", "b", "c"]
    const envInitialValue: Env[] = []
    const wrapper = mount(
      <Formik
        initialValues={{
          [groupNameFieldSpec.name]: groupNameInitialValue,
          [imageFieldSpec.name]: imageInitialValue,
          [commandFieldSpec.name]: commandInitialValue,
          [memoryFieldSpec.name]: memoryInitialValue,
          [cpuFieldSpec.name]: cpuInitialValue,
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
    expect(formGroups).toHaveLength(6)
    expect(fields).toHaveLength(6)
    expect(wrapper.find(EnvFieldArray)).toHaveLength(1)
    expect(wrapper.find(FieldError)).toHaveLength(0)

    // Group name field.
    const groupNameFieldIndex = 0
    expect(formGroups.at(groupNameFieldIndex).props().label).toEqual(
      groupNameFieldSpec.label
    )
    expect(formGroups.at(groupNameFieldIndex).props().helperText).toEqual(
      groupNameFieldSpec.description
    )
    expect(fields.at(groupNameFieldIndex).props().name).toEqual(
      groupNameFieldSpec.name
    )
    expect(fields.at(groupNameFieldIndex).props().value).toEqual(
      groupNameInitialValue
    )

    // Image field.
    const imageFieldIndex = 1
    expect(formGroups.at(imageFieldIndex).props().label).toEqual(
      imageFieldSpec.label
    )
    expect(formGroups.at(imageFieldIndex).props().helperText).toEqual(
      imageFieldSpec.description
    )
    expect(fields.at(imageFieldIndex).props().name).toEqual(imageFieldSpec.name)
    expect(
      fields
        .at(imageFieldIndex)
        .find("input")
        .props().value
    ).toEqual(imageInitialValue)

    // Command field.
    const commandFieldIndex = 2
    expect(formGroups.at(commandFieldIndex).props().label).toEqual(
      commandFieldSpec.label
    )
    expect(formGroups.at(commandFieldIndex).props().helperText).toEqual(
      commandFieldSpec.description
    )
    expect(fields.at(commandFieldIndex).props().name).toEqual(
      commandFieldSpec.name
    )
    expect(
      fields
        .at(commandFieldIndex)
        .find("textarea")
        .props().value
    ).toEqual(commandInitialValue)

    // CPU field.
    const cpuFieldIndex = 3
    expect(formGroups.at(cpuFieldIndex).props().label).toEqual(
      cpuFieldSpec.label
    )
    expect(formGroups.at(cpuFieldIndex).props().helperText).toEqual(
      cpuFieldSpec.description
    )
    expect(fields.at(cpuFieldIndex).props().name).toEqual(cpuFieldSpec.name)
    expect(
      fields
        .at(cpuFieldIndex)
        .find("input")
        .props().value
    ).toEqual(cpuInitialValue)

    // Memory field.
    const memoryFieldIndex = 4
    expect(formGroups.at(memoryFieldIndex).props().label).toEqual(
      memoryFieldSpec.label
    )
    expect(formGroups.at(memoryFieldIndex).props().helperText).toEqual(
      memoryFieldSpec.description
    )
    expect(fields.at(memoryFieldIndex).props().name).toEqual(
      memoryFieldSpec.name
    )
    expect(
      fields
        .at(memoryFieldIndex)
        .find("input")
        .props().value
    ).toEqual(memoryInitialValue)

    // Tags field.
    const tagsFieldIndex = 5
    expect(formGroups.at(tagsFieldIndex).props().label).toEqual(
      tagsFieldSpec.label
    )
    expect(formGroups.at(tagsFieldIndex).props().helperText).toEqual(
      tagsFieldSpec.description
    )
    expect(fields.at(tagsFieldIndex).props().name).toEqual(tagsFieldSpec.name)
    expect(fields.at(tagsFieldIndex).props().value).toEqual(tagsInitialValue)
  })
})
