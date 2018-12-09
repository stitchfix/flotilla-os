import { stringToSelectOpt, selectOptToString } from "../reactSelectHelpers"
import { IReactSelectOption } from "../../.."

describe("reactSelectHelpers", () => {
  describe("stringToSelectOpt", () => {
    it("transforms a string into a react-select option", () => {
      const s = "some_string"
      expect(stringToSelectOpt(s)).toEqual({
        label: s,
        value: s,
      })
    })
  })

  describe("selectOptToString", () => {
    it("transforms a react-select option into a string", () => {
      const opt: IReactSelectOption = {
        label: "opt",
        value: "opt",
      }
      expect(selectOptToString(opt)).toEqual(opt.value)
    })
  })
})
