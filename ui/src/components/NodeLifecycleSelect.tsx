import * as React from "react"
import Select from "react-select"
import { SelectOption, SelectProps, NodeLifecycle } from "../types"
import * as helpers from "../helpers/selectHelpers"

/**
 * NodeLifecycleSelect
 */
export const NodeLifecycleSelect: React.FunctionComponent<SelectProps & {
  options: SelectOption[]
}> = props => {
  return (
    <Select<SelectOption>
      value={helpers.stringToSelectOpt(props.value)}
      options={[
        { label: NodeLifecycle.SPOT, value: NodeLifecycle.SPOT },
        { label: NodeLifecycle.ON_DEMAND, value: NodeLifecycle.ON_DEMAND },
      ]}
      isClearable
      onChange={option => {
        props.onChange(helpers.preprocessSelectOption(option))
      }}
      styles={helpers.selectStyles}
      theme={helpers.selectTheme}
      isDisabled={props.isDisabled}
    />
  )
}

export default NodeLifecycleSelect
