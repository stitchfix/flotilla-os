import * as React from "react"
import { MemoryRouter } from "react-router-dom"
import { mount } from "enzyme"
import ConnectedDeleteTaskButton, {
  DeleteTaskButton,
  Props,
} from "../DeleteTaskButton"
import Request, { RequestStatus } from "../Request"
import api from "../../api"

jest.mock("../../helpers/FlotillaClient")

const defaultProps: Props = {
  requestStatus: RequestStatus.NOT_READY,
  data: null,
  isLoading: false,
  error: null,
  request: jest.fn(),
  definitionID: "definitionID",
  receivedAt: new Date(),
}

describe("DeleteTaskButton", () => {
  it("calls props.request with the correct args when this.handleSubmitClick is called", () => {
    const r = jest.fn()
    const wrapper = mount<DeleteTaskButton>(
      <DeleteTaskButton {...defaultProps} request={r} />
    )
    expect(r).toHaveBeenCalledTimes(0)
    wrapper.instance().handleSubmitClick()
    expect(r).toHaveBeenCalledTimes(1)
    expect(r).toHaveBeenCalledWith({
      definitionID: wrapper.prop("definitionID"),
    })
  })

  it("provides api.deleteTask as the requestFn", () => {
    // Note: this is testing the connected component so it must be wrapper in
    // a MemoryRouter component.
    const wrapper = mount(
      <MemoryRouter>
        <ConnectedDeleteTaskButton definitionID="id" />
      </MemoryRouter>
    )
    expect(wrapper.find(Request).prop("requestFn")).toEqual(api.deleteTask)
  })
})
