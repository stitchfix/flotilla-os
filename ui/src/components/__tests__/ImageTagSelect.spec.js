import React from "react"
import { configureSetup } from "../../__testutils__"
import { ImageTagSelect } from "../ImageTagSelect"

const setup = configureSetup({
  unconnected: ImageTagSelect,
})

describe("ImageTagSelect", () => {
  describe("Lifecycle methods", () => {
    let fetch = ImageTagSelect.prototype.fetchTags
    beforeEach(() => {
      ImageTagSelect.prototype.fetchTags = jest.fn()
    })
    afterEach(() => {
      ImageTagSelect.prototype.fetchTags = fetch
    })
    it("calls fetchTags when the component mounts", () => {
      const image = "some_image"
      const wrapper = setup({
        connectToReduxForm: true,
        formName: "test",
        props: { image },
      })

      expect(ImageTagSelect.prototype.fetchTags).toHaveBeenCalledTimes(1)
      expect(ImageTagSelect.prototype.fetchTags).toHaveBeenCalledWith(image)
    })
  })
  describe("Non-lifecycle methods", () => {
    it("renders a ReduxFormGroupSelect component", () => {
      const wrapper = setup({
        connectToReduxForm: true,
        formName: "test",
      })

      expect(wrapper.find("ReduxFormGroupSelect").length).toBe(1)
    })
  })
})
