import React from "react"
import { reducer as formReducer } from "redux-form"
import { createStore, combineReducers } from "redux"
import { configureSetup, generateTaskRes } from "../../__testutils__"
import envNameValueDelimiterChar from "../../constants/envNameValueDelimiterChar"
import ConnectedRunForm, { RunForm } from "../RunForm"

const definitionId = "definitionId"
const taskDefinition = generateTaskRes(definitionId)
const store = createStore(combineReducers({ form: formReducer }))
const setup = configureSetup({
  connected: ConnectedRunForm,
  unconnected: RunForm,
})

describe("RunForm", () => {
  const consoleError = console.error
  beforeAll(() => {
    console.error = jest.fn()
  })
  afterAll(() => {
    console.error = consoleError
  })
  it("calls setQuery appropriately", () => {
    const setQuery = RunForm.prototype.setQuery
    RunForm.prototype.setQuery = jest.fn()

    // Scenario A: data has been fetched and passed down from it's parent
    // component (<TaskContainer>) and the router query is empty. This usually
    // happens when you click into the <RunForm> from the
    // <TaskDefinitionView>'s "Run" button.
    const wrapperA = setup({
      props: {
        data: taskDefinition,
        query: {},
      },
      shallow: true,
    })

    // When the component mounts, `didSetQuery` should be false.
    expect(wrapperA.state().didSetQuery).toBeFalsy()

    // Since the props meet the criteria for setting the query, call setQuery.
    expect(wrapperA.instance().shouldSetQuery()).toBe(true)
    expect(RunForm.prototype.setQuery).toHaveBeenCalledTimes(1)

    // Clear the mock call count.
    RunForm.prototype.setQuery.mockClear()

    // Scenario B: data has _not_ been fetched and the router query is empty.
    // This happens when you navigate directly to `/tasks/:id/run` in the
    // browser. In this scenario, setQuery will not be called when the
    // component mounts but when the component updates (after completing the
    // request for the task definition).
    const wrapperB = setup({
      props: {
        data: undefined,
        query: {},
      },
      shallow: true,
    })

    // No data yet, shouldn't set query.
    expect(wrapperB.instance().shouldSetQuery()).toBe(false)
    expect(RunForm.prototype.setQuery).toHaveBeenCalledTimes(0)

    // Mock receiving data from server.
    wrapperB.setProps({ data: taskDefinition })

    // NOW it should set query.
    expect(wrapperB.instance().shouldSetQuery()).toBe(true)
    expect(RunForm.prototype.setQuery).toHaveBeenCalledTimes(1)

    // Clear the mock call count.
    RunForm.prototype.setQuery.mockClear()

    // Scenario C: data has not been fetched but the router query is populated.
    // This can happen when 1) clicking a run's Retry <Link> button, which adds
    // all the environment variables used in the run to the Link's `to` props'
    // `query` key (<Link to={{ pathname: "...", query: run.env.map(...) }})),
    // or 2) when navigating to the RunForm's url with the query already
    // populated. In THIS scenario, setQuery should *not* be called (since
    // there is an existing query).
    const clusterName = "my_cluster"
    const wrapperC = setup({
      props: {
        data: undefined,
        query: {
          cluster: clusterName,
          env: taskDefinition.env.map(
            e => `${e.name}${envNameValueDelimiterChar}${e.value}`
          ),
        },
      },
      shallow: true,
    })

    // No data yet, shouldn't set query.
    expect(wrapperC.instance().shouldSetQuery()).toBe(false)
    expect(RunForm.prototype.setQuery).toHaveBeenCalledTimes(0)

    // Mock receiving data from server.
    wrapperC.setProps({ data: taskDefinition })

    // Even after receiving data, setQuery shouldn't be called since there is
    // an existing query.
    expect(wrapperC.instance().shouldSetQuery()).toBe(false)
    expect(RunForm.prototype.setQuery).toHaveBeenCalledTimes(0)

    // // Clear the mock call count.
    RunForm.prototype.setQuery.mockClear()

    // Restore mock.
    RunForm.prototype.setQuery = setQuery
  })
})
