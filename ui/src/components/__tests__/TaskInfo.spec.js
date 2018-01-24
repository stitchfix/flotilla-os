import React from "react"
import moment from "moment"
import { configureSetup, generateTaskRes } from "../../__testutils__"
import TaskInfo from "../TaskInfo"

const definitionId = "some_id"
const res = generateTaskRes(definitionId)
const setup = configureSetup({
  unconnected: TaskInfo,
  baseProps: { data: res },
})

describe("TaskInfo", () => {
  it("renders the correct run metadata", () => {
    const wrapper = setup()
    // Required metadata + number of env vars
    const numOfFormGroups = 9 + res.env.length
    const formGroups = wrapper.find("FormGroup")

    expect(wrapper.find("Card").length).toEqual(2)
    expect(formGroups.length).toEqual(numOfFormGroups)

    expect(formGroups.at(0).props().label).toEqual("Alias")
    expect(formGroups.at(0).props().children).toEqual(res.alias)

    expect(formGroups.at(1).props().label).toEqual("Definition ID")
    expect(formGroups.at(1).props().children).toEqual(res.definition_id)

    expect(formGroups.at(2).props().label).toEqual("Container Name")
    expect(formGroups.at(2).props().children).toEqual(res.container_name)

    expect(formGroups.at(3).props().label).toEqual("Group Name")
    expect(formGroups.at(3).props().children).toEqual(res.group_name)

    expect(formGroups.at(4).props().label).toEqual("Image")
    expect(formGroups.at(4).props().children).toEqual(res.image)

    expect(formGroups.at(5).props().label).toEqual("Command")
    expect(
      formGroups
        .at(5)
        .find("pre")
        .text()
    ).toMatch(res.command)

    expect(formGroups.at(6).props().label).toEqual("Memory")
    expect(formGroups.at(6).props().children).toEqual(res.memory)

    expect(formGroups.at(7).props().label).toEqual("Arn")
    expect(formGroups.at(7).props().children).toEqual(res.arn)

    expect(formGroups.at(8).props().label).toEqual("Tags")
    expect(formGroups.at(8).find(".pl-tag").length).toEqual(res.tags.length)

    expect(formGroups.at(9).text()).toMatch(res.env[0].name)
    expect(formGroups.at(9).text()).toMatch(res.env[0].value)
  })
})
