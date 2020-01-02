import React from "react"
import { mount } from "enzyme"
import Creatable from "react-select/lib/Creatable"
import Connected, { ClusterSelect } from "../ClusterSelect"
import api from "../../api"

jest.mock("../../helpers/FlotillaClient")

describe("ClusterSelect", () => {
  describe("Unconnected", () => {
    it("renders a Creatable component", () => {
      const props = {
        options: [
          { label: "a", value: "a" },
          { label: "b", value: "b" },
          { label: "c", value: "c" },
        ],
        value: "a",
        onChange: jest.fn(),
      }
      const wrapper = mount(<ClusterSelect {...props} isDisabled={false} />)
      const select = wrapper.find(Creatable)

      // Ensure <Select> component is rendered.
      expect(select).toHaveLength(1)

      // Ensure <Select> component has correct `options` prop.
      expect(select.prop("options")).toEqual(props.options)

      // Ensure <Select> component has correct `value` prop.
      expect(select.prop("value")).toEqual({
        label: props.value,
        value: props.value,
      })

      // Ensure props.onChange is called when <Select>'s onChange prop is
      // called.
      expect(props.onChange).toHaveBeenCalledTimes(0)
      const onChangeProp = select.prop("onChange")
      if (onChangeProp) {
        onChangeProp({ label: "b", value: "b" }, { action: "select-option" })
      }
      expect(props.onChange).toHaveBeenCalledTimes(1)
    })
  })

  describe("Connected", () => {
    beforeEach(() => {
      jest.clearAllMocks()
    })

    it("calls api.listClusters", () => {
      expect(api.listClusters).toHaveBeenCalledTimes(0)
      mount(<Connected value="" onChange={jest.fn()} isDisabled={false} />)
      expect(api.listClusters).toHaveBeenCalledTimes(1)
    })

    it("sends an empty array to the select if the server returns null", () => {
      const mk = jest.spyOn(api, "listClusters")
      mk.mockImplementationOnce(
        () =>
          new Promise(resolve => {
            resolve({
              offset: 0,
              limit: 10,
              clusters: null,
              total: 0,
            })
          })
      )
      const wrapper = mount(
        <Connected value="" onChange={jest.fn()} isDisabled={false} />
      )
      const unconnected = wrapper.find(ClusterSelect)
      expect(unconnected).toHaveLength(1)
      expect(unconnected.prop("options")).toEqual([])
    })
  })
})
