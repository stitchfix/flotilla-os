import moment from "moment"
import { has, get } from "lodash"

const addZeroPadding = str => (str.toString().length === 1 ? `0${str}` : str)
const formatDuration = ({ days, hours, minutes, seconds }) => {
  let duration = ``
  if (days) {
    duration += `${days}.`
  }
  if (hours) {
    duration += `${addZeroPadding(hours)}:`
  } else {
    duration += `00:`
  }
  if (minutes) {
    duration += `${addZeroPadding(minutes)}:`
  } else {
    duration += `00:`
  }
  duration += `${addZeroPadding(seconds)}`
  return duration
}

export default function getRunDuration(run) {
  if (has(run, "started_at")) {
    const start = new Date(run.started_at)
    const end = has(run, "finished_at") ? new Date(run.finished_at) : new Date()
    const diff = end - start
    return formatDuration(moment.duration(diff)._data)
  }
  return "-"
}
