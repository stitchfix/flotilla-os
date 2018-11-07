import invalidRunEnv from "../../constants/invalidRunEnv"
import envNameValueDelimiterChar from "../../constants/envNameValueDelimiterChar"
import getRetryEnv from "../getRetryEnv"

describe("getRetryEnv", () => {
  it("filters out invalid run environment variables and maps the environment variables to a string", () => {
    const env = [
      { name: "foo", value: "bar" },
      { name: "foo2", value: "bar2" },
      // This one should be filtered out.
      { name: invalidRunEnv[0], value: "blarg" },
    ]

    expect(getRetryEnv(env)).toEqual([
      `foo${envNameValueDelimiterChar}bar`,
      `foo2${envNameValueDelimiterChar}bar2`,
    ])
  })
})
