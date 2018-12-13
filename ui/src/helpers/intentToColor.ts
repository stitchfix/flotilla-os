import colors from "./colors"
import { flotillaUIIntents } from "../.."

const intentToColor = (intent?: flotillaUIIntents): string => {
  switch (intent) {
    case flotillaUIIntents.PRIMARY:
      return colors.blue[0]
    case flotillaUIIntents.ERROR:
      return colors.red[0]
    case flotillaUIIntents.WARNING:
      return colors.yellow[0]
    case flotillaUIIntents.SUCCESS:
      return colors.green[0]
    case flotillaUIIntents.SUBTLE:
      return colors.light_gray[2]
    default:
      return colors.black[4]
  }
}

export default intentToColor
