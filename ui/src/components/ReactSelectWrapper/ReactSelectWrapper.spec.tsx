import * as React from "react"
import { mount } from "enzyme"
import Select from "react-select"
import CreatableSelect from "react-select/lib/Creatable"
import { ValueType } from "react-select/lib/types"
import { ReactSelectWrapper, IProps } from "./ReactSelectWrapper"
import { flotillaUIRequestStates, IReactSelectOption } from "../../types"
import {
  stringToSelectOpt,
  selectOptToString,
} from "../../helpers/reactSelectHelpers"

const DEFAULT_PROPS: IProps = {
  error: false,
  inFlight: false,
  isCreatable: false,
  isMulti: false,
  name: "select",
  onChange: (value: string | string[]) => {},
  onRequestError: (e: any) => {},
  options: [],
  request: (args?: any) => {},
  requestState: flotillaUIRequestStates.NOT_READY,
  shouldRequestOptions: false,
  value: "",
}

describe("ReactSelectWrapper", () => {
  describe("Lifecycle Methods", () => {
    it("calls props.request when the component mounts if necessary", () => {
      const request = jest.fn()
      mount(
        <ReactSelectWrapper
          {...DEFAULT_PROPS}
          request={request}
          shouldRequestOptions={false}
        />
      )
      expect(request).toHaveBeenCalledTimes(0)

      mount(
        <ReactSelectWrapper
          {...DEFAULT_PROPS}
          request={request}
          shouldRequestOptions
        />
      )
      expect(request).toHaveBeenCalledTimes(1)
    })
  })

  describe("getValue", () => {
    it("returns the correct value for multi-selects if the raw value is an array", () => {
      const value = ["one", "two", "three"]
      const expected = value.map(stringToSelectOpt)
      const wrapper = mount(
        <ReactSelectWrapper {...DEFAULT_PROPS} isMulti value={value} />
      )
      const inst = wrapper.instance() as ReactSelectWrapper
      expect(inst.getValue()).toEqual(expected)
    })
    it("returns the correct value for multi-selects if the raw value is a string", () => {
      const value = "one"
      const expected = [stringToSelectOpt(value)]
      const wrapper = mount(
        <ReactSelectWrapper {...DEFAULT_PROPS} isMulti value={value} />
      )
      const inst = wrapper.instance() as ReactSelectWrapper
      expect(inst.getValue()).toEqual(expected)
    })
    it("returns the correct value for multi-selects if the raw value is empty", () => {
      const value = null
      const expected: any[] = []
      const wrapper = mount(
        <ReactSelectWrapper {...DEFAULT_PROPS} isMulti value={value} />
      )
      const inst = wrapper.instance() as ReactSelectWrapper
      expect(inst.getValue()).toEqual(expected)
    })
    it("returns the correct value for single-selects", () => {
      const value = "hello"
      const expected = stringToSelectOpt(value)
      const wrapper = mount(
        <ReactSelectWrapper {...DEFAULT_PROPS} value={value} />
      )
      const inst = wrapper.instance() as ReactSelectWrapper
      expect(inst.getValue()).toEqual(expected)
    })
  })

  describe("handleSelectChange", () => {
    it("calls props.onChange with an empty array for multi-selects if the selected value is null or undefined", () => {
      const onChange = jest.fn()
      const selectedValue = null
      const expected: any[] = []
      const wrapper = mount(
        <ReactSelectWrapper {...DEFAULT_PROPS} isMulti onChange={onChange} />
      )
      const inst = wrapper.instance() as ReactSelectWrapper
      inst.handleSelectChange(selectedValue)
      expect(onChange).toHaveBeenCalledTimes(1)
      expect(onChange).toHaveBeenCalledWith(expected)
    })
    it("calls props.onChange with an array of strings for multi-selects", () => {
      const onChange = jest.fn()
      const selectedValue: ValueType<IReactSelectOption[]> = [
        { label: "foo", value: "foo" },
        { label: "bar", value: "bar" },
      ]
      const expected = selectedValue.map(selectOptToString)
      const wrapper = mount(
        <ReactSelectWrapper {...DEFAULT_PROPS} isMulti onChange={onChange} />
      )
      const inst = wrapper.instance() as ReactSelectWrapper
      inst.handleSelectChange(selectedValue)
      expect(onChange).toHaveBeenCalledTimes(1)
      expect(onChange).toHaveBeenCalledWith(expected)
    })
    it("calls props.onChange with a blank string for single-selects if the selected value is null or undefined", () => {
      const onChange = jest.fn()
      const selectedValue = null
      const expected = ""
      const wrapper = mount(
        <ReactSelectWrapper {...DEFAULT_PROPS} onChange={onChange} />
      )
      const inst = wrapper.instance() as ReactSelectWrapper
      inst.handleSelectChange(selectedValue)
      expect(onChange).toHaveBeenCalledTimes(1)
      expect(onChange).toHaveBeenCalledWith(expected)
    })
    it("calls props.onChange with string for single-selects", () => {
      const onChange = jest.fn()
      const selectedValue: ValueType<IReactSelectOption> = {
        label: "foo",
        value: "foo",
      }
      const expected = selectOptToString(selectedValue)
      const wrapper = mount(
        <ReactSelectWrapper {...DEFAULT_PROPS} onChange={onChange} />
      )
      const inst = wrapper.instance() as ReactSelectWrapper
      inst.handleSelectChange(selectedValue)
      expect(onChange).toHaveBeenCalledTimes(1)
      expect(onChange).toHaveBeenCalledWith(expected)
    })
  })

  describe("isReady", () => {
    it("returns false if props.shouldRequestOptions is true and options haven't been fetched", () => {
      const wrapper = mount(
        <ReactSelectWrapper
          {...DEFAULT_PROPS}
          shouldRequestOptions
          requestState={flotillaUIRequestStates.NOT_READY}
        />
      )
      const inst = wrapper.instance() as ReactSelectWrapper
      expect(inst.isReady()).toEqual(false)
    })
    it("returns true if props.shouldRequestOptions is true and options have been fetched", () => {
      const wrapper = mount(
        <ReactSelectWrapper
          {...DEFAULT_PROPS}
          shouldRequestOptions
          requestState={flotillaUIRequestStates.READY}
        />
      )
      const inst = wrapper.instance() as ReactSelectWrapper
      expect(inst.isReady()).toEqual(true)
    })
    it("returns true if props.shouldRequestOptions is false", () => {
      const wrapper = mount(
        <ReactSelectWrapper
          {...DEFAULT_PROPS}
          shouldRequestOptions={false}
          requestState={flotillaUIRequestStates.NOT_READY}
        />
      )
      const inst = wrapper.instance() as ReactSelectWrapper
      expect(inst.isReady()).toEqual(true)
    })
  })

  describe("render", () => {
    it("renders a CreatableSelect if props.isCreatable is true", () => {
      const wrapper = mount(
        <ReactSelectWrapper {...DEFAULT_PROPS} isCreatable />
      )
      expect(wrapper.find(CreatableSelect).length).toEqual(1)
      expect(wrapper.find(Select).length).toEqual(0)
    })
    it("renders a Select if props.isCreatable is false", () => {
      const wrapper = mount(
        <ReactSelectWrapper {...DEFAULT_PROPS} isCreatable={false} />
      )
      expect(wrapper.find(CreatableSelect).length).toEqual(0)
      expect(wrapper.find(Select).length).toEqual(1)
    })
  })
})
