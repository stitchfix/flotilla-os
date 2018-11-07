import qs from "qs"

export default function queryObjectToSearchString(query) {
  return `?${qs.stringify(query)}`
}
