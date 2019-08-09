import FlotillaClient from "./helpers/FlotillaClient"

const client = new FlotillaClient({
  baseURL: process.env.REACT_APP_BASE_URL || "FILL_ME_IN",
})

export default client
