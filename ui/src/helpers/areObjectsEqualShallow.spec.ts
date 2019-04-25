import areObjectsEqualShallow from "./areObjectsEqualShallow"

describe("areObjectsEqualShallow", () => {
  it("returns false if either arguments are not objects", () => {
    expect(areObjectsEqualShallow({}, [])).toEqual(false)
  })
  it("returns false if objects aren't the same size", () => {
    expect(areObjectsEqualShallow({ foo: "bar" }, {})).toEqual(false)
  })
  it("compares the stringfied value for the same key of each object", () => {
    expect(
      areObjectsEqualShallow({ a: 1, b: 2, c: 3 }, { a: 1, b: 2, c: "3" })
    ).toEqual(true)
    expect(
      areObjectsEqualShallow({ a: 1, b: 2, c: 3 }, { a: 1, b: 2, c: 4 })
    ).toEqual(false)
  })
  it("returns true if the objects are equal", () => {
    expect(areObjectsEqualShallow({ a: 1, b: 2 }, { a: 1, b: 2 })).toEqual(true)
  })
})
