import getImageTagsEndpoint from "../getImageTagsEndpoint"

describe("getImageTagsEndpoint", () => {
  it("parses the string correctly", () => {
    const strStart = "foo"
    const strEnd = "bar"
    const str = `${strStart}{image}${strEnd}`
    const image = "my_image"

    expect(getImageTagsEndpoint(str, image)).toEqual(
      `${strStart}${image}${strEnd}`
    )
  })
})
