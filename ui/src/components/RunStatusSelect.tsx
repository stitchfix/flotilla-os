import * as React from "react"
import { isArray } from "lodash"
import Select from "react-select"
import { SelectOption, MultiSelectProps, RunStatus } from "../types"
import * as helpers from "../helpers/selectHelpers"

const RunStatusSelect: React.FunctionComponent<MultiSelectProps> = props => {
  let v: SelectOption[]
  if (!isArray(props.value)) {
    v = [helpers.stringToSelectOpt(props.value)]
  } else {
    v = props.value.map(helpers.stringToSelectOpt)
  }
  return (
    <Select<SelectOption>
      value={v}
      options={[
        { label: RunStatus.PENDING, value: RunStatus.PENDING },
        { label: RunStatus.QUEUED, value: RunStatus.QUEUED },
        { label: RunStatus.RUNNING, value: RunStatus.RUNNING },
      ]}
      onChange={option => {
        props.onChange(helpers.preprocessMultiSelectOption(option))
      }}
      isMulti
      styles={helpers.selectStyles}
      theme={helpers.selectTheme}
      isDisabled={props.isDisabled}
    />
  )
}

export default RunStatusSelect
