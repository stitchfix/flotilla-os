import React from "react"
import { MemoryRouter } from "react-router-dom"
import { shallow } from "enzyme"
import configureMockStore from "redux-mock-store"
import thunk from "redux-thunk"
import axios from "axios"
import axiosMockAdapter from "axios-mock-adapter"
import TaskDefinitionView from "../TaskDefinitionView"
import { generateTaskRes, configureSetup } from "../../__testutils__"

const axiosMock = new axiosMockAdapter(axios)
axiosMock.onGet().reply(200)
const setup = configureSetup({
  connected: TaskDefinitionView,
})
const middlewares = [thunk]
const mockStore = configureMockStore(middlewares)

describe("TaskDefinitionView", () => {
  it("renders the correct children", () => {
    const wrapper = setup({
      connectToRedux: true,
      connectToRouter: true,
      store: mockStore({}),
    })
    expect(wrapper.find("View").length).toBe(1)
    expect(wrapper.find("ViewHeader").length).toBe(1)
    expect(wrapper.find("TaskInfo").length).toBe(1)
    expect(wrapper.find("TaskHistory").length).toBe(1)
  })
})
