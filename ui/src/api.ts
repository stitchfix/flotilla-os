import FlotillaClient from "./helpers/FlotillaClient"

let baseURL: string

if (process.env.NODE_ENV === "production" && process.env.REACT_APP_BASE_URL) {
  baseURL = process.env.REACT_APP_BASE_URL
} else if (
  process.env.NODE_ENV === "development" &&
  process.env.REACT_APP_BASE_URL_DEV
) {
  baseURL = process.env.REACT_APP_BASE_URL_DEV
} else {
  throw new Error(
    "Base URL undefined. If you are running this in development, please set the `REACT_APP_BASE_URL_DEV` environment variable. If you are running this in production, please set the `REACT_APP_BASE_URL` environment variable."
  )
}

const client = new FlotillaClient({ baseURL })

export default client
