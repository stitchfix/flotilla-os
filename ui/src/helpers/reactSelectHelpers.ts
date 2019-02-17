import { get } from "lodash"
import { IReactSelectOption } from "../types"
import colors from "./colors"
import { Theme } from "react-select/lib/types"

export const preprocessSelectValue = (selected: IReactSelectOption): string => {
  if (selected === null || selected === undefined) return ""

  return selected.value
}
export const preprocessMultiSelectValue = (
  selected: IReactSelectOption[]
): string[] => {
  if (selected === null || selected === undefined) return []

  return selected.map(selectOptToString)
}

export const stringToSelectOpt = (str: string): IReactSelectOption => ({
  label: str,
  value: str,
})

export const selectOptToString = (opt: IReactSelectOption): string =>
  get(opt, "value", "")

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
    color: colors.light_gray[1],
  }),
  option: (provided: any) => ({
    ...provided,
    color: colors.gray[3],
    paddingTop: 12,
    paddingBottom: 12,
  }),
}

export const selectTheme = (theme: Theme): Theme => ({
  ...theme,
  colors: {
    ...theme.colors,
    primary: colors.blue[0],
    primary75: colors.blue[0],
    primary50: colors.blue[0],
    primary25: colors.black[4],
    danger: colors.red[0],
    dangerLight: colors.red[4],
    neutral0: colors.black[1],
    neutral5: colors.black[1],
    neutral10: colors.black[4],
    neutral20: colors.black[4],
    neutral30: colors.black[4],
    neutral40: colors.gray[0],
    neutral50: colors.gray[1],
    neutral60: colors.gray[2],
    neutral70: colors.gray[3],
    neutral80: colors.gray[4],
    neutral90: colors.light_gray[0],
  },
})
