import React from "react"
import { MemoryRouter } from "react-router-dom"
import { Provider } from "react-redux"
import configureMockStore from "redux-mock-store"
import thunk from "redux-thunk"
import axios from "axios"
import axiosMockAdapter from "axios-mock-adapter"
import { modalActions, popupActions } from "aa-ui-components"
import { configureSetup } from "../../__testutils__"
import config from "../../config"
import ConnectedStopRunModal, { StopRunModal } from "../StopRunModal"

const definitionId = "definitionId"
const runId = "runId"
const axiosMock = new axiosMockAdapter(axios)
const middlewares = [thunk]
const mockStore = configureMockStore(middlewares)
const setup = configureSetup({
  connected: ConnectedStopRunModal,
  unconnected: StopRunModal,
  baseProps: { definitionId, runId },
})

describe("StopRunModal", () => {
  afterEach(() => {
    axiosMock.reset()
  })
  it("renders 1 Cancel and 1 Delete button", () => {
    const wrapper = setup()
    const buttons = wrapper.find("Button")
    expect(buttons.length).toBe(2)
    expect(buttons.at(0).text()).toBe("Cancel")
    expect(buttons.at(1).text()).toBe("Stop Run")
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
      .onDelete(`${config.FLOTILLA_API}/task/${definitionId}/history/${runId}`)
      .reply(200, {
        deleted: true,
      })
    const dispatch = jest.fn()
    const wrapper = setup({
      props: {
        dispatch,
      },
    })
    wrapper
      .instance()
      .handleStopButtonClick()
      .then(() => {
        // Dispatch two actions - unrenderModal and renderPopup
        expect(dispatch).toHaveBeenCalledTimes(2)
        expect(dispatch).toHaveBeenCalledWith(
          popupActions.renderPopup(expect.anything())
        )
        expect(dispatch).toHaveBeenCalledWith(modalActions.unrenderModal())
      })
  })
  it("handles failed requests", () => {
    axiosMock
      .onDelete(`${config.FLOTILLA_API}/task/${definitionId}/history/${runId}`)
      .reply(500, {})
    const dispatch = jest.fn()
    const wrapper = setup({
      props: {
        dispatch,
      },
    })
    wrapper
      .instance()
      .handleStopButtonClick()
      .then(() => {
        // Dispatch one action - render a popup w/ failure message.
        expect(dispatch).toHaveBeenCalledTimes(1)
        expect(dispatch).toHaveBeenCalledWith(
          popupActions.renderPopup(expect.anything())
        )
      })
  })
})
