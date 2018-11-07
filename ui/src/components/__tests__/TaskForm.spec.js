import React from "react"
import { reducer as formReducer } from "redux-form"
import { createStore, combineReducers } from "redux"
import { get } from "lodash"
import { configureSetup, generateTaskRes } from "../../__testutils__"
import taskFormTypes from "../../constants/taskFormTypes"
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
  let editForm
  let createForm
  let copyForm
  const consoleError = console.error
  beforeAll(() => {
    console.error = jest.fn()
    editForm = setup(createSetupOpts(taskFormTypes.edit))
    createForm = setup(createSetupOpts(taskFormTypes.create))
    copyForm = setup(createSetupOpts(taskFormTypes.copy))
  })
  afterAll(() => {
    console.error = consoleError
  })
  it("renders the correct title", () => {
    const editFormComponent = editForm.find("TaskForm")
    expect(editForm.find("ViewHeader").props().title).toEqual(
      `Edit ${get(editFormComponent.props().data, "alias", definitionId)}`
    )

    const createFormComponent = createForm.find("TaskForm")
    expect(createForm.find("ViewHeader").props().title).toEqual(
      "Create New Task"
    )

    const copyFormComponent = copyForm.find("TaskForm")
    expect(copyForm.find("ViewHeader").props().title).toEqual(
      `Copy ${get(copyFormComponent.props().data, "alias", definitionId)}`
    )
  })
  it("renders the correct fields", () => {
    const fields = createForm.find("Field")

    expect(createForm.find("EnvFieldArray").length).toBe(1)
    expect(fields.length).toEqual(6)
    expect(fields.at(0).props().name).toEqual("alias")
    expect(fields.at(1).props().name).toEqual("group_name")
    expect(fields.at(2).props().name).toEqual("image")
    expect(fields.at(3).props().name).toEqual("command")
    expect(fields.at(4).props().name).toEqual("memory")
    expect(fields.at(5).props().name).toEqual("tags")
  })
  it("doesn't render an `alias` field when editing a task", () => {
    // Note: all the "redux-form-helper" components in `aa-ui-components`
    // ultimately render a <Field>, which is why we can `.find` it this way.
    expect(editForm.find("Field").length).toEqual(5)
    expect(createForm.find("Field").length).toEqual(6)
    expect(copyForm.find("Field").length).toEqual(6)
  })
})
