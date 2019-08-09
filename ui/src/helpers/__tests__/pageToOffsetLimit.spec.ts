import pageToOffsetLimit from "../pageToOffsetLimit"

describe("pageToOffsetLimit", () => {
  it("works correctly", () => {
    expect(pageToOffsetLimit({ page: 1, limit: 20 })).toEqual({
      offset: 0,
      limit: 20,
    })
    expect(pageToOffsetLimit({ page: 2, limit: 20 })).toEqual({
      offset: 20,
      limit: 20,
    })
  })
})
