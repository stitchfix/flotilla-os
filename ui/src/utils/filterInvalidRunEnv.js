import config from "../config"

/**
 * Helper function to filter invalid runtime environment variables.
 */
const filterInvalidRunEnv = (env = []) =>
  env.filter(e => !config.INVALID_RUN_ENV.includes(e.name))

export default filterInvalidRunEnv
