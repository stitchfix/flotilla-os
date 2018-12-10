import * as React from "react"
import { mount } from "enzyme"
import { MemoryRouter } from "react-router-dom"
import DeleteTaskModal from "../DeleteTaskModal"
import api from "../../../api"

describe("DeleteTaskModal", () => {
  it("provides the correct requestFn and getRequestArgs props to ConfirmModal", () => {
    const definitionID = "definitionID"
    const runID = "runID"
    const wrapper = mount(
      <MemoryRouter>
        <DeleteTaskModal definitionID={definitionID} />
      </MemoryRouter>
    )
    const confirmModal = wrapper.find("ConfirmModal")
    expect(confirmModal.prop("requestFn")).toEqual(api.deleteTask)

    // Note: need to cast as function in order to call.
    const getRequestArgsProp = confirmModal.prop("getRequestArgs") as Function
    expect(getRequestArgsProp()).toEqual({ definitionID })
  })
})
