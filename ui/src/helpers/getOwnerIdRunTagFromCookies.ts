import { get, has } from "lodash"
import cookie from "cookie"

const getOwnerIdRunTagFromCookies = (): string => {
  // Get owner ID.
  let ownerID: string = "flotilla-ui"

  // Check if the `REACT_APP_RUN_TAG_OWNER_ID_COOKIE_PATH` env var is set;
  // proceed to extract it from the cookies if so.
  if (process.env.REACT_APP_RUN_TAG_OWNER_ID_COOKIE_PATH) {
    const cookies = cookie.parse(document.cookie)
    const cookiePath = process.env.REACT_APP_RUN_TAG_OWNER_ID_COOKIE_PATH.split(
      "."
    )

    if (cookiePath.length > 1 && has(cookies, cookiePath[0])) {
      ownerID = get(
        JSON.parse(get(cookies, cookiePath[0], "{}")),
        cookiePath.slice(1),
        "flotilla-ui"
      )
    } else {
      ownerID = get(
        cookies,
        process.env.REACT_APP_RUN_TAG_OWNER_ID_COOKIE_PATH,
        "flotilla-ui"
      )
    }
  }

  return ownerID
}

export default getOwnerIdRunTagFromCookies
