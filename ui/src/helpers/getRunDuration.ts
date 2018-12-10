import * as moment from "moment"
import { IFlotillaRun } from "../.."

const addZeroPadding = (value: number): string =>
  value.toString().length === 1 ? `0${value}` : value.toString()

/**
 * Given a run object, this function returns a human readable timestamp for the
 * run duration.
 *
 * @param run - the Run object
 * @returns a human readable timestamp
 */
const getRunDuration = (run: Partial<IFlotillaRun>): string => {
  if (!!run.started_at) {
    const start: number = new Date(run.started_at).valueOf()
    const end: number = !!run.finished_at
      ? new Date(run.finished_at).valueOf()
      : new Date().valueOf()
    const duration = moment.duration(end - start)
    const months = duration.get("months")
    const days = duration.get("days")
    const hours = duration.get("hours")
    const minutes = duration.get("minutes")
    const seconds = duration.get("seconds")

    let ret = ""

    if (months !== 0) {
      ret += `${months} months, `
    }

    if (days !== 0) {
      ret += `${days} days, `
    }

    ret += `${addZeroPadding(hours)}:${addZeroPadding(
      minutes
    )}:${addZeroPadding(seconds)}`

    return ret
  }

  return "-"
}

export default getRunDuration
