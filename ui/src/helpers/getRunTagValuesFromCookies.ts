import { getJSON } from "js-cookie"
import { get } from "lodash"
import config from "../config"

const getRunTagValuesFromCookies = (): { [run_tag_key: string]: any } => {
  let ret: any = {}
  for (let i = 0; i < config.COOKIES_TO_RUN_TAGS.length; i++) {
    const [key, cookiePath] = config.COOKIES_TO_RUN_TAGS[i].split("|")
    const splitCookiePath = cookiePath.split(".")

    if (splitCookiePath.length === 0) continue
    if (splitCookiePath.length === 1) {
      ret[key] = getJSON(splitCookiePath[0])
    } else {
      ret[key] = get(
        getJSON(splitCookiePath[0]),
        splitCookiePath.slice(1, splitCookiePath.length),
        ""
      )
    }
  }

  return ret
}

export default getRunTagValuesFromCookies
