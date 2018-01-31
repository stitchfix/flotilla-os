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
  beforeAll(() => {
    console.warn = jest.fn()
  })
  afterAll(() => {
    console.warn = warn
  })
  it("renders the correct run metadata", () => {
    const wrapper = setup({ connectToRouter: true })
    // Required metadata + number of env vars
    const numOfFormGroups = 10 + res.env.length
    const formGroups = wrapper.find("FormGroup")

    expect(wrapper.find("Card").length).toEqual(2)
    expect(formGroups.length).toEqual(numOfFormGroups)

    expect(formGroups.at(0).props().label).toEqual("Cluster")
    expect(formGroups.at(0).props().children).toEqual(res.cluster)

    expect(formGroups.at(1).props().label).toEqual("Exit Code")
    expect(formGroups.at(1).props().children).toEqual(res.exit_code)

    expect(formGroups.at(2).props().label).toEqual("Started At")
    expect(formGroups.at(2).text()).toMatch(moment(res.started_at).fromNow())

    expect(formGroups.at(3).props().label).toEqual("Finished At")
    expect(formGroups.at(3).text()).toMatch(moment(res.finished_at).fromNow())

    expect(formGroups.at(4).props().label).toEqual("Run ID")
    expect(formGroups.at(4).find("Link")).toBeTruthy()
    expect(
      formGroups
        .at(4)
        .find("Link")
        .text()
    ).toEqual(res.run_id)
    expect(
      formGroups
        .at(4)
        .find("Link")
        .props().to
    ).toEqual(`/runs/${res.run_id}`)

    expect(formGroups.at(5).props().label).toEqual("Task Definition ID")
    expect(formGroups.at(5).text()).toMatch(res.definition_id)

    expect(formGroups.at(6).props().label).toEqual("Image")
    expect(formGroups.at(6).text()).toMatch(res.image)

    expect(formGroups.at(7).props().label).toEqual("Task Arn")
    expect(formGroups.at(7).text()).toMatch(res.task_arn)

    expect(formGroups.at(8).props().label).toEqual("Instance ID")
    expect(formGroups.at(8).props().children).toEqual(res.instance.instance_id)

    expect(formGroups.at(9).props().label).toEqual("Instance DNS Name")
    expect(formGroups.at(9).props().children).toEqual(res.instance.dns_name)

    expect(formGroups.at(10).text()).toMatch(res.env[0].name)
    expect(formGroups.at(10).text()).toMatch(res.env[0].value)
  })
})
