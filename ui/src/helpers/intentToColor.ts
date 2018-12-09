import colors from "./colors"
import { intents } from "../.."

const intentToColor = (intent: intents): string => {
  switch (intent) {
    case intents.PRIMARY:
      return colors.blue[0]
    case intents.ERROR:
      return colors.red[0]
    case intents.WARNING:
      return colors.yellow[0]
    case intents.SUCCESS:
      return colors.green[0]
    case intents.SUBTLE:
      return colors.light_gray[2]
    default:
      return colors.black[4]
  }
}

export default intentToColor
