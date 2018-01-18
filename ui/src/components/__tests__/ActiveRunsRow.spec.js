import React from "react"
import { mount } from "enzyme"
import { MemoryRouter } from "react-router-dom"
import { generateRunRes } from "../../__testutils__"
import ActiveRunsRow from "../ActiveRunsRow"

const RUN_ID = "RUN_ID"
const res = generateRunRes(RUN_ID)

describe("ActiveRunsRow", () => {
  const onStopButtonClick = jest.fn()
  const wrapper = mount(
    <MemoryRouter>
      <ActiveRunsRow data={res} onStopButtonClick={onStopButtonClick} />
    </MemoryRouter>
  )
  it("renders a Link to /runs/:run_id", () => {
    expect(wrapper.find("Link").length).toBe(1)
    expect(wrapper.find("Link").props().to).toEqual(`/runs/${RUN_ID}`)
  })
  it("renders 5 table cells with the correct flex lengths", () => {
    expect(wrapper.find(".pl-td").length).toEqual(5)
    expect(
      wrapper
        .find(".pl-td")
        .at(0)
        .props().style
    ).toMatchObject({
      flex: 1,
    })
    expect(
      wrapper
        .find(".pl-td")
        .at(1)
        .props().style
    ).toMatchObject({
      flex: 1,
    })
    expect(
      wrapper
        .find(".pl-td")
        .at(2)
        .props().style
    ).toMatchObject({
      flex: 1.5,
    })
    expect(
      wrapper
        .find(".pl-td")
        .at(3)
        .props().style
    ).toMatchObject({
      flex: 4,
    })
    expect(
      wrapper
        .find(".pl-td")
        .at(4)
        .props().style
    ).toMatchObject({
      flex: 1.5,
    })
  })
  it("calls props.onStopButtonClick when clicking on the Stop Button", () => {
    wrapper.find("Button").simulate("click")
    expect(onStopButtonClick).toHaveBeenCalledTimes(1)
  })
})
