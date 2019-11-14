import * as React from "react"
import { isArray } from "lodash"
import Creatable from "react-select/lib/Creatable"
import { SelectOption, MultiSelectProps } from "../types"
import * as helpers from "../helpers/selectHelpers"

const GenericMultiSelect: React.FunctionComponent<MultiSelectProps> = props => {
  let value = props.value
  if (!isArray(props.value)) {
    value = [props.value]
  }

  return (
    <Creatable<SelectOption>
      value={value.map(helpers.stringToSelectOpt)}
      options={[]}
      onChange={option => {
        props.onChange(helpers.preprocessMultiSelectOption(option))
      }}
      isMulti
      isClearable
      styles={helpers.selectStyles}
      theme={helpers.selectTheme}
      isDisabled={props.isDisabled}
    />
  )
}

export default GenericMultiSelect
