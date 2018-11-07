import React from "react"
import axios from "axios"
import axiosMockAdapter from "axios-mock-adapter"
import TaskDefinitionView from "../TaskDefinitionView"
import { configureSetup } from "../../__testutils__"

const axiosMock = new axiosMockAdapter(axios)
axiosMock.onGet().reply(200)
const setup = configureSetup({
  connected: TaskDefinitionView,
})

describe("TaskDefinitionView", () => {
  const error = console.error
  beforeAll(() => {
    console.error = jest.fn()
  })
  afterAll(() => {
    console.error = error
  })
  it("renders the correct children", () => {
    const wrapper = setup({
      connectToRedux: true,
      connectToRouter: true,
    })
    expect(wrapper.find("View").length).toBe(1)
    expect(wrapper.find("ViewHeader").length).toBe(1)
    expect(wrapper.find("TaskInfo").length).toBe(1)
    expect(wrapper.find("TaskHistory").length).toBe(1)
  })
})
