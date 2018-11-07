import qs from "qs"

export default function searchStringToQueryObject(str) {
  return qs.parse(str.replace("?", ""))
}
