import * as React from "react"
import { FastField } from "formik"
import CreatableSelect from "react-select/lib/Creatable"
import { IReactSelectOption } from "../../types"
import StyledField from "../styled/Field"
import {
  stringToSelectOpt,
  preprocessSelectValue,
  preprocessMultiSelectValue,
} from "../../helpers/reactSelectHelpers"

export const AliasField: React.SFC<{}> = () => (
  <StyledField
    label="Alias"
    description="Choose a descriptive alias for this task."
    isRequired
  >
    <FastField name="alias" />
  </StyledField>
)

export const CommandField: React.SFC<{}> = () => (
  <StyledField
    label="Command"
    description="The command for this task to execute."
    isRequired
  >
    <FastField name="command" component="textarea" />
  </StyledField>
)

export const GroupNameField: React.SFC<{
  onChange: (value: string) => void
  value: string
  options: IReactSelectOption[]
}> = ({ onChange, value, options }) => (
  <StyledField
    label="Group Name"
    description="Create a new group name or select an existing one to help searching for this task in the future."
    isRequired
  >
    <FastField
      name="group_name"
      onChange={(selected: IReactSelectOption) => {
        onChange(preprocessSelectValue(selected))
      }}
      value={stringToSelectOpt(value)}
      component={CreatableSelect}
      options={options}
    />
  </StyledField>
)

export const ImageField: React.SFC<{}> = () => (
  <StyledField
    label="Image"
    description="The full URL of the Docker image and tag."
    isRequired
  >
    <FastField name="image" />
  </StyledField>
)

export const MemoryField: React.SFC<{}> = () => (
  <StyledField
    label="Memory (MB)"
    description="The amount of memory this task needs."
    isRequired
  >
    <FastField name="memory" type="number" />
  </StyledField>
)

export const TagsField: React.SFC<{
  onChange: (value: string[]) => void
  value: string[]
  options: IReactSelectOption[]
}> = ({ onChange, value, options }) => (
  <StyledField label="Tags">
    <FastField
      name="tags"
      onChange={(selected: IReactSelectOption[]) => {
        onChange(preprocessMultiSelectValue(selected))
      }}
      value={value.map(stringToSelectOpt)}
      component={CreatableSelect}
      options={options}
      isMulti
    />
  </StyledField>
)
