import { isArray } from "lodash"
import { SelectOption } from "../types"
import { ValueType, Theme } from "react-select/lib/types"
import { Colors } from "@blueprintjs/core"

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

export const selectStyles = {
  container: (provided: any) => ({
    ...provided,
    width: "100%",
  }),
  control: (provided: any) => ({
    ...provided,
    borderWidth: 2,
  }),
  menu: (provided: any) => ({
    ...provided,
    color: Colors.LIGHT_GRAY1,
  }),
  option: (provided: any) => ({
    ...provided,
    color: Colors.LIGHT_GRAY1,
    paddingTop: 8,
    paddingBottom: 8,
  }),
}

export const selectTheme = (theme: Theme): Theme => ({
  ...theme,
  colors: {
    ...theme.colors,
    primary: Colors.COBALT1,
    primary75: Colors.COBALT1,
    primary50: Colors.COBALT1,
    primary25: Colors.COBALT1,
    danger: Colors.RED1,
    dangerLight: Colors.RED4,
    neutral0: Colors.BLACK,
    neutral5: Colors.BLACK,
    neutral10: Colors.DARK_GRAY4,
    neutral20: Colors.DARK_GRAY4,
    neutral30: Colors.DARK_GRAY4,
    neutral40: Colors.GRAY1,
    neutral50: Colors.GRAY1,
    neutral60: Colors.GRAY2,
    neutral70: Colors.GRAY3,
    neutral80: Colors.GRAY4,
    neutral90: Colors.LIGHT_GRAY1,
  },
})
