import axios, { AxiosError } from "axios"
import { get } from "lodash"
import AxiosMockAdapter from "axios-mock-adapter"
import * as urljoin from "url-join"
import * as qs from "qs"
import FlotillaAPIClient from "../FlotillaAPIClient"
import {
  IFlotillaCreateTaskPayload,
  IFlotillaRunTaskPayload,
  IFlotillaEditTaskPayload,
} from "../../.."

const ROOT_LOCATION = "ROOT_LOCATION"
const SUCCESS_PATH = "/mock_path"
const ERROR_PATH = "/error_path"
const MOCK_RESPONSE = { foo: "bar" }
const MOCK_ERROR = new Error("Request failed with status code 404")
const MOCK_ERROR_STATUS_CODE = 500
const mock = new AxiosMockAdapter(axios)

mock.onGet(`${ROOT_LOCATION}${SUCCESS_PATH}`).reply(() => {
  return [200, MOCK_RESPONSE]
})

mock.onGet(`${ROOT_LOCATION}${ERROR_PATH}`).reply(() => {
  return [MOCK_ERROR_STATUS_CODE, MOCK_ERROR]
})

describe("FlotillaAPIClient", () => {
  describe("Initialization", () => {
    it("sets the `location` member", () => {
      const api = new FlotillaAPIClient(ROOT_LOCATION)
      expect(api.location).toBe(ROOT_LOCATION)
    })
  })

  describe("Base Methods", () => {
    describe("constructURL", () => {
      it("appends the `path` to the `location` member", () => {
        const api = new FlotillaAPIClient(ROOT_LOCATION)
        const path = "some_path/some_other_path"
        const url = api.constructURL({ path })
        expect(url).toBe(urljoin(ROOT_LOCATION, path))
      })

      it("appends a stringified query to the end if the query is set", () => {
        const api = new FlotillaAPIClient(ROOT_LOCATION)
        const path = "some_path/some_other_path"
        const query = { a: 1, b: 2, c: 3 }
        const url = api.constructURL({ path, query })
        expect(url).toBe(
          `${urljoin(ROOT_LOCATION, path)}?${qs.stringify(query)}`
        )
      })
    })

    describe("request", () => {
      it("resolves with the response's data if the request is successful", () => {
        const api = new FlotillaAPIClient(ROOT_LOCATION)
        return expect(
          api.request({ method: "get", path: SUCCESS_PATH })
        ).resolves.toEqual(MOCK_RESPONSE)
      })

      it("rejects if the request is unsuccessful", () => {
        const api = new FlotillaAPIClient(ROOT_LOCATION)
        return expect(
          api.request({ method: "get", path: ERROR_PATH })
        ).rejects.toEqual(expect.anything())
      })
    })

    describe("processError", () => {
      const baseError = {
        config: {},
        name: "name",
        message: "message",
      }

      it("handles response errors", () => {
        const responseError: AxiosError = {
          ...baseError,
          response: {
            data: "error",
            status: 500,
            statusText: "",
            headers: "",
            config: {},
          },
        }
        const api = new FlotillaAPIClient(ROOT_LOCATION)
        expect(api.processError(responseError)).toEqual({
          data: get(responseError, ["message"]),
          status: get(responseError, ["response", "status"]),
          headers: get(responseError, ["response", "headers"]),
        })
      })

      it("handles request errors", () => {
        const requestError: AxiosError = {
          ...baseError,
          request: "request error message",
        }
        const api = new FlotillaAPIClient(ROOT_LOCATION)
        expect(api.processError(requestError)).toEqual({
          data: get(requestError, "request"),
        })
      })

      it("handles other errors", () => {
        const otherError: AxiosError = {
          ...baseError,
        }
        const api = new FlotillaAPIClient(ROOT_LOCATION)
        expect(api.processError(otherError)).toEqual({
          data: baseError.message,
        })
      })
    })
  })

  describe("Flotilla Methods", () => {
    const api = new FlotillaAPIClient(ROOT_LOCATION)
    const mockRequest = jest.spyOn(api, "request")
    mockRequest.mockImplementation(() => {})

    afterEach(() => {
      mockRequest.mockClear()
    })

    it("getTasks method calls request method with the correct arguments", () => {
      expect(api.request).toHaveBeenCalledTimes(0)
      api.getTasks({ query: {} })
      expect(api.request).toHaveBeenCalledTimes(1)
      expect(api.request).toHaveBeenCalledWith({
        method: "get",
        path: "/v1/task",
        query: {},
      })
    })

    it("getTask method calls request method with the correct arguments", () => {
      const definitionID = "definitionID"
      expect(api.request).toHaveBeenCalledTimes(0)
      api.getTask({ definitionID })
      expect(api.request).toHaveBeenCalledTimes(1)
      expect(api.request).toHaveBeenCalledWith({
        method: "get",
        path: `/v1/task/${definitionID}`,
      })
    })

    it("getTaskByAlias method calls request method with the correct arguments", () => {
      const alias = "alias"
      expect(api.request).toHaveBeenCalledTimes(0)
      api.getTaskByAlias({ alias })
      expect(api.request).toHaveBeenCalledTimes(1)
      expect(api.request).toHaveBeenCalledWith({
        method: "get",
        path: `/v1/task/alias/${alias}`,
      })
    })

    it("getTaskHistory method calls request method with the correct arguments", () => {
      const definitionID = "definitionID"
      const query = {}
      expect(api.request).toHaveBeenCalledTimes(0)
      api.getTaskHistory({ definitionID, query })
      expect(api.request).toHaveBeenCalledTimes(1)
      expect(api.request).toHaveBeenCalledWith({
        method: "get",
        path: `/v1/task/${definitionID}/history`,
        query,
      })
    })

    it("createTask method calls request method with the correct arguments", () => {
      const values: IFlotillaCreateTaskPayload = {
        memory: 1024,
        image: "",
        group_name: "",
        command: "",
        alias: "",
      }
      expect(api.request).toHaveBeenCalledTimes(0)
      api.createTask({ values })
      expect(api.request).toHaveBeenCalledTimes(1)
      expect(api.request).toHaveBeenCalledWith({
        method: "post",
        path: `/v1/task`,
        payload: values,
      })
    })

    it("updateTask method calls request method with the correct arguments", () => {
      const definitionID = "definitionID"
      const values: IFlotillaEditTaskPayload = {
        memory: 1024,
        image: "",
        group_name: "",
        command: "",
      }
      expect(api.request).toHaveBeenCalledTimes(0)
      api.updateTask({ definitionID, values })
      expect(api.request).toHaveBeenCalledTimes(1)
      expect(api.request).toHaveBeenCalledWith({
        method: "put",
        path: `/v1/task/${definitionID}`,
        payload: values,
      })
    })

    it("deleteTask method calls request method with the correct arguments", () => {
      const definitionID = "definitionID"

      expect(api.request).toHaveBeenCalledTimes(0)
      api.deleteTask({ definitionID })
      expect(api.request).toHaveBeenCalledTimes(1)
      expect(api.request).toHaveBeenCalledWith({
        method: "delete",
        path: `/v1/task/${definitionID}`,
      })
    })

    it("runTask method calls request method with the correct arguments", () => {
      const definitionID = "definitionID"
      const values: IFlotillaRunTaskPayload = {
        cluster: "cluster",
      }
      expect(api.request).toHaveBeenCalledTimes(0)
      api.runTask({ definitionID, values })
      expect(api.request).toHaveBeenCalledTimes(1)
      expect(api.request).toHaveBeenCalledWith({
        method: "put",
        path: `/v4/task/${definitionID}/execute`,
        payload: values,
      })
    })

    it("getActiveRuns method calls request method with the correct arguments", () => {
      const query = {}
      expect(api.request).toHaveBeenCalledTimes(0)
      api.getActiveRuns({ query })
      expect(api.request).toHaveBeenCalledTimes(1)
      expect(api.request).toHaveBeenCalledWith({
        method: "get",
        path: "/v1/history",
        query,
      })
    })

    it("stopRun method calls request method with the correct arguments", () => {
      const definitionID = "definitionID"
      const runID = "runID"
      expect(api.request).toHaveBeenCalledTimes(0)
      api.stopRun({ definitionID, runID })
      expect(api.request).toHaveBeenCalledTimes(1)
      expect(api.request).toHaveBeenCalledWith({
        method: "delete",
        path: `/v1/task/${definitionID}/history/${runID}`,
      })
    })

    it("getRunLogs method calls request method with the correct arguments", () => {
      const runID = "runID"
      const lastSeen = "lastSeen"
      expect(api.request).toHaveBeenCalledTimes(0)
      api.getRunLogs({ runID, lastSeen })
      expect(api.request).toHaveBeenCalledTimes(1)
      expect(api.request).toHaveBeenCalledWith({
        method: "get",
        path: `/v1/${runID}/logs`,
        query: { last_seen: lastSeen },
      })
    })

    it("getGroups method calls request method with the correct arguments", () => {
      expect(api.request).toHaveBeenCalledTimes(0)
      api.getGroups()
      expect(api.request).toHaveBeenCalledTimes(1)
      expect(api.request).toHaveBeenCalledWith({
        method: "get",
        path: `/v1/groups`,
        query: { limit: 2000 },
        preprocess: expect.any(Function),
      })
    })

    it("getClusters method calls request method with the correct arguments", () => {
      expect(api.request).toHaveBeenCalledTimes(0)
      api.getClusters()
      expect(api.request).toHaveBeenCalledTimes(1)
      expect(api.request).toHaveBeenCalledWith({
        method: "get",
        path: `/v1/clusters`,
        preprocess: expect.any(Function),
      })
    })

    it("getTags method calls request method with the correct arguments", () => {
      expect(api.request).toHaveBeenCalledTimes(0)
      api.getTags()
      expect(api.request).toHaveBeenCalledTimes(1)
      expect(api.request).toHaveBeenCalledWith({
        method: "get",
        path: `/v1/tags`,
        query: { limit: 5000 },
        preprocess: expect.any(Function),
      })
    })
  })
})
