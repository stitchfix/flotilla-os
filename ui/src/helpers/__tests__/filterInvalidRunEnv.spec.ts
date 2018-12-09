import filterInvalidRunEnv from "../filterInvalidRunEnv"
import { IFlotillaEnv } from "../../.."
import config from "../../config"

describe("filterInvalidRunEnv", () => {
  it("filters out invalid env vars", () => {
    config.INVALID_RUN_ENV = ["foo"]
    const unfiltered: IFlotillaEnv[] = [
      { name: "foo", value: "foo" },
      { name: "bar", value: "bar" },
      { name: "baz", value: "baz" },
    ]
    const filtered: IFlotillaEnv[] = [
      { name: "bar", value: "bar" },
      { name: "baz", value: "baz" },
    ]
    expect(filterInvalidRunEnv(unfiltered)).toEqual(filtered)
  })
})
