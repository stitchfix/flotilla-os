import React from "react"
import configureMockStore from "redux-mock-store"
import thunk from "redux-thunk"
import axios from "axios"
import axiosMockAdapter from "axios-mock-adapter"
import modalActions from "../../actions/modalActions"
import popupActions from "../../actions/popupActions"
import { configureSetup } from "../../__testutils__"
import config from "../../config"
import ConnectedDeleteTaskModal, { DeleteTaskModal } from "../DeleteTaskModal"

const definitionId = "definitionId"
const axiosMock = new axiosMockAdapter(axios)
const middlewares = [thunk]
const mockStore = configureMockStore(middlewares)
const setup = configureSetup({
  connected: ConnectedDeleteTaskModal,
  unconnected: DeleteTaskModal,
  baseProps: { definitionId },
})

describe("DeleteTaskModal", () => {
  afterEach(() => {
    axiosMock.reset()
  })
  it("renders 1 Cancel and 1 Delete button", () => {
    const wrapper = setup()
    const buttons = wrapper.find("Button")
    expect(buttons.length).toBe(2)
    expect(buttons.at(0).text()).toBe("Cancel")
    expect(buttons.at(1).text()).toBe("Delete Task")
  })
  it("disables the Delete button if the request is in flight", () => {
    const wrapper = setup()
    wrapper.setState({
      inFlight: true,
    })
    expect(
      wrapper
        .find("Button")
        .at(1)
        .props().isLoading
    ).toBe(true)
  })
  it("handles successful requests", () => {
    axiosMock
      .onDelete(`${config.FLOTILLA_API}/task/${definitionId}`)
      .reply(200, {
        deleted: true,
      })
    const dispatch = jest.fn()
    const push = jest.fn()
    const wrapper = setup({
      props: {
        dispatch,
        history: { push },
      },
    })
    wrapper
      .instance()
      .handleDeleteButtonClick()
      .then(() => {
        // Dispatch two actions - unrenderModal and renderPopup
        expect(dispatch).toHaveBeenCalledTimes(2)
        expect(dispatch).toHaveBeenCalledWith(
          popupActions.renderPopup(expect.anything())
        )
        expect(dispatch).toHaveBeenCalledWith(modalActions.unrenderModal())

        // Push to /tasks
        expect(push).toHaveBeenCalledTimes(1)
        expect(push).toHaveBeenCalledWith("/tasks")
      })
  })
  it("handles failed requests", () => {
    axiosMock
      .onDelete(`${config.FLOTILLA_API}/task/${definitionId}`)
      .reply(500, {})
    const dispatch = jest.fn()
    const push = jest.fn()
    const wrapper = setup({
      props: {
        dispatch,
        history: { push },
      },
    })
    wrapper
      .instance()
      .handleDeleteButtonClick()
      .then(() => {
        // Dispatch one action - render a popup w/ failure message.
        expect(dispatch).toHaveBeenCalledTimes(1)
        expect(dispatch).toHaveBeenCalledWith(
          popupActions.renderPopup(expect.anything())
        )
      })
  })
})
