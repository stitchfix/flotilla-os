import config from "./config"
import FlotillaAPIClient from "./FlotillaAPIClient"

const api = new FlotillaAPIClient({ location: config.FLOTILLA_API })

export default api
