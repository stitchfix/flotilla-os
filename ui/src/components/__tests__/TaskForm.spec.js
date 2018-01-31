import React from "react"
import { mount } from "enzyme"
import { MemoryRouter } from "react-router-dom"
// Note: need to create an actual store w/ form reducer.
import { reducer as formReducer } from "redux-form"
import { createStore, combineReducers } from "redux"
import { configureSetup, generateTaskRes } from "../../__testutils__"
import { taskFormTypes } from "../../constants/"
import ConnectedTaskForm, { TaskForm } from "../TaskForm"

const definitionId = "definitionId"
const taskDefinition = generateTaskRes(definitionId)
const setup = configureSetup({
  baseProps: {
    data: taskDefinition,
  },
  connected: ConnectedTaskForm,
  unconnected: TaskForm,
})

const createSetupOpts = (taskFormType, props) => ({
  props: { taskFormType, ...props },
  connectToRouter: true,
  connectToReduxForm: true,
  formName: "task",
  store: createStore(combineReducers({ form: formReducer })),
})

const sharedSetupOpts = {}

describe("TaskForm", () => {
  const consoleError = console.error
  beforeAll(() => {
    console.error = jest.fn()
  })
  afterAll(() => {
    console.error = consoleError
  })
  it("renders the correct title", () => {})
  it("doesn't render an `alias` field when editing a task", () => {
    const editForm = setup(createSetupOpts(taskFormTypes.edit))

    // Note: all the "redux-form-helper" components in `aa-ui-components`
    // ultimately render a <Field>, which is why we can `.find` it this way.
    expect(editForm.find("Field").length).toEqual(5)

    const createForm = setup(createSetupOpts(taskFormTypes.create))
    expect(createForm.find("Field").length).toEqual(6)

    const copyForm = setup(createSetupOpts(taskFormTypes.copy))
    expect(copyForm.find("Field").length).toEqual(6)
  })
})
