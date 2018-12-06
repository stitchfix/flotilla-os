import colors from "./colors"
import intentTypes from "./intentTypes"

const intentToColor = intent => {
  switch (intent) {
    case intentTypes.primary:
      return colors.blue[0]
    case intentTypes.error:
      return colors.red[0]
    case intentTypes.warning:
      return colors.yellow[0]
    case intentTypes.success:
      return colors.green[0]
    case intentTypes.subtle:
      return colors.light_gray[2]
    default:
      return colors.black[4]
  }
}

export default intentToColor
