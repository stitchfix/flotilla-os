import { invalidRunEnv, envNameValueDelimiterChar } from "../constants"

export default function getRetryEnv(env) {
  return env
    .filter(e => !invalidRunEnv.includes(e.name))
    .map(e => `${e.name}${envNameValueDelimiterChar}${e.value}`)
}
