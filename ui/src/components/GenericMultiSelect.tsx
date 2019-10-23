import * as React from "react"
import Creatable from "react-select/lib/Creatable"
import { SelectOption, MultiSelectProps } from "../types"
import * as helpers from "../helpers/selectHelpers"

const GenericMultiSelect: React.FunctionComponent<MultiSelectProps> = props => {
  return (
    <Creatable<SelectOption>
      value={props.value.map(helpers.stringToSelectOpt)}
      options={[]}
      onChange={option => {
        props.onChange(helpers.preprocessMultiSelectOption(option))
      }}
      isMulti
      isClearable
      styles={helpers.selectStyles}
      theme={helpers.selectTheme}
    />
  )
}

export default GenericMultiSelect
