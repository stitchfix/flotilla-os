import React from "react"
import { reducer as formReducer } from "redux-form"
import { createStore, combineReducers } from "redux"
import { configureSetup } from "../../__testutils__"
import EnvFieldArray from "../EnvFieldArray"

const setup = configureSetup({
  connected: EnvFieldArray,
  unconnected: EnvFieldArray,
})

// @TODO: finish this.
describe("EnvFieldArray", () => {
  const consoleError = console.error
  beforeAll(() => {
    console.error = jest.fn()
  })
  afterAll(() => {
    console.error = consoleError
  })
  it("works", () => {
    const wrapper = setup({
      connectToRouter: true,
      connectToReduxForm: true,
      formName: "test",
      store: createStore(combineReducers({ form: formReducer })),
    })
  })
})
