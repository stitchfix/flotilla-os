import invalidRunEnv from "../constants/invalidRunEnv"
import envNameValueDelimiterChar from "../constants/envNameValueDelimiterChar"

export default function getRetryEnv(env) {
  return env
    .filter(e => !invalidRunEnv.includes(e.name))
    .map(e => `${e.name}${envNameValueDelimiterChar}${e.value}`)
}
