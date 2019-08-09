import * as React from "react"
import { MemoryRouter } from "react-router-dom"
import { mount } from "enzyme"
import ConnectedStopRunButton, { StopRunButton, Props } from "../StopRunButton"
import Request, { RequestStatus } from "../Request"
import api from "../../api"

const defaultProps: Props = {
  requestStatus: RequestStatus.NOT_READY,
  data: null,
  isLoading: false,
  error: null,
  request: jest.fn(),
  definitionID: "definitionID",
  runID: "runID",
}

describe("StopRunButton", () => {
  it("calls props.request with the correct args when this.handleSubmitClick is called", () => {
    const r = jest.fn()
    const wrapper = mount<StopRunButton>(
      <StopRunButton {...defaultProps} request={r} />
    )
    expect(r).toHaveBeenCalledTimes(0)
    wrapper.instance().handleSubmitClick()
    expect(r).toHaveBeenCalledTimes(1)
    expect(r).toHaveBeenCalledWith({
      definitionID: wrapper.prop("definitionID"),
      runID: wrapper.prop("runID"),
    })
  })

  it("provides api.stopRun as the requestFn", () => {
    // Note: this is testing the connected component so it must be wrapper in
    // a MemoryRouter component.
    const wrapper = mount(
      <MemoryRouter>
        <ConnectedStopRunButton definitionID="id" runID="rid" />
      </MemoryRouter>
    )
    expect(wrapper.find(Request).prop("requestFn")).toEqual(api.stopRun)
  })
})
