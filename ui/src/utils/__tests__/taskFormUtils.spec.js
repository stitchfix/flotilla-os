import { generateTaskRes } from "../../__testutils__"
import config from "../../config"
import { taskFormTypes } from "../../constants/"
import * as taskFormUtils from "../taskFormUtils"

describe("taskFormUtils", () => {
  describe("mapStateToProps", () => {
    it("outputs the correct structure", () => {
      const definitionId = "definitionId"
      const data = generateTaskRes(definitionId)
      const reduxState = {
        selectOpts: { inFlight: true },
      }
      const ownProps = {
        taskFormType: taskFormTypes.create,
      }
      const output = taskFormUtils.mapStateToProps(reduxState, ownProps)
      expect(output).toHaveProperty("selectOptionsRequestInFlight")
      expect(output).toHaveProperty("groupOptions")
      expect(output).toHaveProperty("tagOptions")
    })
    it("adds initialValues if props.taskFormType is copy or edit", () => {
      const definitionId = "definitionId"
      const data = generateTaskRes(definitionId)
      const reduxState = {
        selectOpts: { inFlight: true },
      }
      const ownProps = {
        taskFormType: taskFormTypes.edit,
        data,
      }
      expect(
        taskFormUtils.mapStateToProps(reduxState, ownProps)
      ).toHaveProperty("initialValues")
      expect(
        taskFormUtils.mapStateToProps(reduxState, ownProps).initialValues
      ).toEqual({
        group_name: data.group_name,
        command: data.command,
        memory: data.memory,
        env: data.env,
        tags: data.tags,
        image: data.image,
      })
    })
  })
  describe("transformFormValues", () => {
    it("outputs the correct structure")
  })
  describe("validate", () => {
    it("catches errors")
  })
})
