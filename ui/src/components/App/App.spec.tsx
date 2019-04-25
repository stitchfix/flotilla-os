import * as React from "react"
import { BrowserRouter, Route } from "react-router-dom"
import { shallow, ShallowWrapper } from "enzyme"
import App from "./App"
import ModalContainer from "../Modal/ModalContainer"
import PopupContainer from "../Popup/PopupContainer"
import CreateTaskForm from "../TaskForm/CreateTaskForm"
import Tasks from "../Tasks/Tasks"
import ActiveRuns from "../ActiveRuns/ActiveRuns"
import TaskRouter from "../Task/TaskRouter"
import Run from "../Run/Run"

describe("App", () => {
  let wrapper: ShallowWrapper
  beforeAll(() => {
    wrapper = shallow(<App />)
  })

  it("renders a BrowserRouter component", () => {
    expect(wrapper.find(BrowserRouter).length).toBe(1)
  })

  it("renders a ModalContainer component", () => {
    expect(wrapper.find(ModalContainer).length).toBe(1)
  })

  it("renders a PopupContainer component", () => {
    expect(wrapper.find(PopupContainer).length).toBe(1)
  })

  it("renders the correct routes", () => {
    const routes = wrapper.find(Route)
    expect(routes.length).toBe(6)

    const createTaskRoute = routes.at(0)
    expect(createTaskRoute.prop("exact")).toBe(true)
    expect(createTaskRoute.prop("path")).toBe("/tasks/create")
    expect(createTaskRoute.prop("component")).toBe(CreateTaskForm)

    const activeRunsRoute = routes.at(1)
    expect(activeRunsRoute.prop("exact")).toBe(true)
    expect(activeRunsRoute.prop("path")).toBe("/runs")
    expect(activeRunsRoute.prop("component")).toBe(ActiveRuns)

    const tasksRoute = routes.at(2)
    expect(tasksRoute.prop("exact")).toBe(true)
    expect(tasksRoute.prop("path")).toBe("/tasks")
    expect(tasksRoute.prop("component")).toBe(Tasks)

    const taskAliasRoute = routes.at(3)
    expect(taskAliasRoute.prop("path")).toBe("/tasks/alias/:alias")

    const taskRoute = routes.at(4)
    expect(taskRoute.prop("path")).toBe("/tasks/:definitionID")
    expect(taskRoute.prop("component")).toBe(TaskRouter)

    const runRoute = routes.at(5)
    expect(runRoute.prop("path")).toBe("/runs/:runID")
    expect(runRoute.prop("component")).toBe(Run)
  })
})
