import { generateTaskRes } from "../../__testutils__"
import config from "../../config"
import { taskFormTypes } from "../../constants/"
import * as taskFormUtils from "../taskFormUtils"

describe("taskFormUtils", () => {
  describe("joinImage", () => {
    it("joins the image and tag, prefixed with the docker repository host", () => {
      const image = "foo"
      const tag = "bar"
      expect(taskFormUtils.joinImage(image, tag)).toEqual(
        `${config.DOCKER_REPOSITORY_HOST}/${image}:${tag}`
      )
    })
  })
  describe("splitImage", () => {
    it("splits an image string into an object containing an image name and tag", () => {
      const image = "foo"
      const tag = "bar"
      const str = `${config.DOCKER_REPOSITORY_HOST}/${image}:${tag}`
      expect(taskFormUtils.splitImage(str)).toEqual({ image, tag })
    })
  })
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
      expect(output).toHaveProperty("imageOptions")
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
        image: taskFormUtils.splitImage(data.image).image,
        image_tag: taskFormUtils.splitImage(data.image).tag,
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
