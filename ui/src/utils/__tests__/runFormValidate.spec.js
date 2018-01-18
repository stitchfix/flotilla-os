import runFormValidate from "../runFormValidate"

describe("runFormValidate", () => {
  it("catches cluster errors", () => {
    const values = {
      env: [{ name: "foo", value: "bar" }],
    }

    expect(runFormValidate(values)).toHaveProperty("cluster")
    expect(runFormValidate(values)).not.toHaveProperty("env")
  })
  it("catches env errors", () => {
    const values = {
      cluster: "my_cluster",
      env: [{ name: "foo" }],
    }

    expect(runFormValidate(values)).not.toHaveProperty("cluster")
    expect(runFormValidate(values)).toHaveProperty("env")
  })
  it("returns an empty object if no errors", () => {
    const values = {
      cluster: "my_cluster",
      env: [{ name: "foo", value: "bar" }],
    }

    expect(Object.keys(runFormValidate(values)).length).toEqual(0)
  })
})
