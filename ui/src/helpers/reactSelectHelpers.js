import { isString, get } from "lodash"
import colors from "./colors"

export const stringToSelectOpt = (str = "") => {
  let ret = isString(str) ? str : ""
  return { label: ret, value: ret }
}
export const selectOptToString = opt => get(opt, "value", "")
export const selectStyles = {
  container: provided => ({
    ...provided,
    width: "100%",
  }),
  control: provided => ({
    ...provided,
    borderWidth: 2,
  }),
  menu: provided => ({
    ...provided,
    color: colors.light_gray[1],
  }),
  option: provided => ({
    ...provided,
    color: colors.gray[3],
    paddingTop: 12,
    paddingBottom: 12,
  }),
}
export const selectTheme = theme => ({
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
