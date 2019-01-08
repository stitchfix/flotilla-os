import config from "./config"
import FlotillaAPIClient from "./helpers/FlotillaAPIClient"

const api = new FlotillaAPIClient(
  process.env.NODE_ENV === "production"
    ? config.FLOTILLA_API
    : config.FLOTILLA_API_DEV
)

export default api
