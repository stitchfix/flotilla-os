import * as moment from "moment"
import getRunDuration from "../getRunDuration"

describe("getRunDuration", () => {
  it("returns a human-readable timestamp", () => {
    const started_at = "2018-01-31T20:30:45.067Z"
    const finished_at = "2018-01-31T22:26:11.483Z"
    expect(
      getRunDuration({
        started_at,
        finished_at,
      })
    ).toEqual("01:55:26")
  })

  it("returns a `-` char if the run hasn't started", () => {
    expect(getRunDuration({})).toEqual("-")
  })

  it("prepends the number of months and days if necessary", () => {
    const started_at = "2018-01-01T20:30:45.067Z"
    const finished_at = "2018-08-22T22:26:11.483Z"

    expect(
      getRunDuration({
        started_at,
        finished_at,
      })
    ).toEqual("7 months, 19 days, 01:55:26")
  })
})
