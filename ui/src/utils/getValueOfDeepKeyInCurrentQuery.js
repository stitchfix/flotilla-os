import { has } from "lodash"

export default function getValueOfDeepKeyInCurrentQuery(query, key) {
  if (!has(query, key)) {
    return ""
  } else {
    return Array.isArray(query[key]) ? [...query[key]] : [query[key]]
  }
}
