import config from "./config"
import FlotillaAPIClient from "./helpers/FlotillaAPIClient"

const api = new FlotillaAPIClient({ location: config.FLOTILLA_API })

export default api
