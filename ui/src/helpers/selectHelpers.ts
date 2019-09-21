import { isArray } from "lodash"
import { SelectOption } from "../types"
import { ValueType } from "react-select/lib/types"

export const stringToSelectOpt = (s: string): SelectOption => ({
  label: s,
  value: s,
})

export const selectOptToString = (o: SelectOption): string => o.value

export const preprocessSelectOption = (
  option: ValueType<SelectOption>
): string => {
  if (option === null || option === undefined || isArray(option)) return ""
  return option.value
}

export const preprocessMultiSelectOption = (
  options: ValueType<SelectOption>
): string[] => {
  if (options === null || options === undefined || !isArray(options)) return []
  return options.map(selectOptToString)
}
