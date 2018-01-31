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

describe("TaskForm", () => {
  it("doesn't render an `alias` field when editing a task", () => {
    const editForm = setup({
      props: { taskFormType: taskFormTypes.edit },
      connectToRouter: true,
      connectToReduxForm: true,
      formName: "task",
      store: createStore(combineReducers({ form: formReducer })),
    })

    // Note: all the "redux-form-helper" components in `aa-ui-components`
    // ultimately render a <Field>, which is why we can `.find` it this way.
    expect(editForm.find("Field").length).toEqual(5)

    const createForm = setup({
      props: { taskFormType: taskFormTypes.create },
      connectToRouter: true,
      connectToReduxForm: true,
      formName: "task",
      store: createStore(combineReducers({ form: formReducer })),
    })

    // Note: all the "redux-form-helper" components in `aa-ui-components`
    // ultimately render a <Field>, which is why we can `.find` it this way.
    expect(createForm.find("Field").length).toEqual(6)

    const copyForm = setup({
      props: { taskFormType: taskFormTypes.copy },
      connectToRouter: true,
      connectToReduxForm: true,
      formName: "task",
      store: createStore(combineReducers({ form: formReducer })),
    })

    // Note: all the "redux-form-helper" components in `aa-ui-components`
    // ultimately render a <Field>, which is why we can `.find` it this way.
    expect(copyForm.find("Field").length).toEqual(6)
  })
})
