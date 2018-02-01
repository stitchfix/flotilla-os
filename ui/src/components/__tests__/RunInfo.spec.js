import React from "react"
import moment from "moment"
import { configureSetup, generateRunRes } from "../../__testutils__"
import RunInfo from "../RunInfo"

const runId = "some_id"
const res = generateRunRes(runId)
const setup = configureSetup({
  unconnected: RunInfo,
  baseProps: { data: res },
})
const formGroupValueClassName = ".pl-form-group-static"

describe("RunInfo", () => {
  const warn = console.warn
  const error = console.error
  beforeAll(() => {
    console.warn = jest.fn()
    console.error = jest.fn()
  })
  afterAll(() => {
    console.warn = warn
    console.error = error
  })
  it("renders the correct run metadata", () => {
    const wrapper = setup({ connectToRouter: true })
    expect(wrapper.find("Card").length).toEqual(3)

    // Run status (2) + Required metadata (10) + number of env vars
    const numOfRunStatusBarFormGroups = 2
    const numOfRunInfoFormGroups = 10
    const numOfFormGroups =
      numOfRunStatusBarFormGroups + numOfRunInfoFormGroups + res.env.length
    const formGroups = wrapper.find("FormGroup")
    expect(formGroups.length).toEqual(numOfFormGroups)

    expect(
      formGroups.at(numOfRunStatusBarFormGroups + 0).props().label
    ).toEqual("Cluster")
    expect(
      formGroups.at(numOfRunStatusBarFormGroups + 0).props().children
    ).toEqual(res.cluster)

    expect(
      formGroups.at(numOfRunStatusBarFormGroups + 1).props().label
    ).toEqual("Exit Code")
    expect(
      formGroups.at(numOfRunStatusBarFormGroups + 1).props().children
    ).toEqual(res.exit_code)

    expect(
      formGroups.at(numOfRunStatusBarFormGroups + 2).props().label
    ).toEqual("Started At")
    expect(formGroups.at(numOfRunStatusBarFormGroups + 2).text()).toMatch(
      moment(res.started_at).fromNow()
    )

    expect(
      formGroups.at(numOfRunStatusBarFormGroups + 3).props().label
    ).toEqual("Finished At")
    expect(formGroups.at(numOfRunStatusBarFormGroups + 3).text()).toMatch(
      moment(res.finished_at).fromNow()
    )

    expect(
      formGroups.at(numOfRunStatusBarFormGroups + 4).props().label
    ).toEqual("Run ID")
    expect(
      formGroups.at(numOfRunStatusBarFormGroups + 4).find("Link")
    ).toBeTruthy()
    expect(
      formGroups
        .at(numOfRunStatusBarFormGroups + 4)
        .find("Link")
        .text()
    ).toEqual(res.run_id)
    expect(
      formGroups
        .at(numOfRunStatusBarFormGroups + 4)
        .find("Link")
        .props().to
    ).toEqual(`/runs/${res.run_id}`)

    expect(
      formGroups.at(numOfRunStatusBarFormGroups + 5).props().label
    ).toEqual("Task Definition ID")
    expect(formGroups.at(numOfRunStatusBarFormGroups + 5).text()).toMatch(
      res.definition_id
    )

    expect(
      formGroups.at(numOfRunStatusBarFormGroups + 6).props().label
    ).toEqual("Image")
    expect(formGroups.at(numOfRunStatusBarFormGroups + 6).text()).toMatch(
      res.image
    )

    expect(
      formGroups.at(numOfRunStatusBarFormGroups + 7).props().label
    ).toEqual("Task Arn")
    expect(formGroups.at(numOfRunStatusBarFormGroups + 7).text()).toMatch(
      res.task_arn
    )

    expect(
      formGroups.at(numOfRunStatusBarFormGroups + 8).props().label
    ).toEqual("Instance ID")
    expect(
      formGroups.at(numOfRunStatusBarFormGroups + 8).props().children
    ).toEqual(res.instance.instance_id)

    expect(
      formGroups.at(numOfRunStatusBarFormGroups + 9).props().label
    ).toEqual("Instance DNS Name")
    expect(
      formGroups.at(numOfRunStatusBarFormGroups + 9).props().children
    ).toEqual(res.instance.dns_name)

    expect(formGroups.at(numOfRunStatusBarFormGroups + 10).text()).toMatch(
      res.env[0].name
    )
    expect(formGroups.at(numOfRunStatusBarFormGroups + 10).text()).toMatch(
      res.env[0].value
    )
  })
})
