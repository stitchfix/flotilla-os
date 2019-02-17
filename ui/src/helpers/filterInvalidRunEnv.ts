import config from "../config"
import { IFlotillaEnv } from "../types"

const filterInvalidRunEnv = (env: IFlotillaEnv[]): IFlotillaEnv[] =>
  env.filter((e: IFlotillaEnv) => !config.INVALID_RUN_ENV.includes(e.name))

export default filterInvalidRunEnv
