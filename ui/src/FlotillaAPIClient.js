import axios from "axios"
import { isEmpty, isString, isObject, isFunction, get } from "lodash"
import qs from "qs"
import urljoin from "url-join"
import { stringToSelectOpt } from "./utils/reactSelectHelpers"

/**
 * API client to communicate with the Flotilla API
 */
class FlotillaAPIClient {
  constructor({ location }) {
    this.location = location
  }

  getTasks = (query = { offset: 0, limit: 20 }) => {
    return this._request({
      method: "get",
      path: "/v1/task",
      query,
      payload: null,
    })
  }

  getTask = ({ definitionID }) => {
    return this._request({
      method: "get",
      path: `/v1/task/${definitionID}`,
      query: null,
      payload: null,
    })
  }

  getTaskHistory = ({ definitionID, query }) => {
    return this._request({
      method: "get",
      path: `/v1/task/${definitionID}/history`,
      query,
      payload: null,
    })
  }

  getActiveRuns = (query = { offset: 0, limit: 20 }) => {
    return this._request({
      method: "get",
      path: "/v1/history",
      query,
      payload: null,
    })
  }

  createTask = ({ values }) => {
    return this._request({
      method: "post",
      path: "/v1/task",
      query: null,
      payload: values,
    })
  }

  updateTask = ({ definitionID, values }) => {
    return this._request({
      method: "put",
      path: `/v1/task/${definitionID}`,
      query: null,
      payload: values,
    })
  }

  deleteTask = ({ definitionID }) => {
    return this._request({
      method: "delete",
      path: `/v1/task/${definitionID}`,
      query: null,
      payload: null,
    })
  }

  runTask = ({ definitionID, values }) => {
    return this._request({
      method: "put",
      path: `/v4/task/${definitionID}/execute`,
      query: null,
      payload: values,
    })
  }

  stopRun = ({ definitionID, runID }) => {
    return this._request({
      method: "delete",
      path: `/v1/task/${definitionID}/history/${runID}`,
      query: null,
      payload: null,
    })
  }

  getRun = ({ runID }) => {
    return this._request({
      method: "get",
      path: `/v1/task/history/${runID}`,
      query: null,
      payload: null,
    })
  }

  getRunLogs = ({ runID, lastSeen }) => {
    let q = {}

    if (!!lastSeen) {
      q.last_seen = lastSeen
    }

    return this._request({
      method: "get",
      path: `/v1/${runID}/logs`,
      query: q,
      payload: null,
    })
  }

  getGroups = () => {
    return this._request({
      method: "get",
      path: `/v1/groups`,
      query: { limit: 2000 },
      payload: null,
      preprocess: res =>
        get(res, "groups", [])
          .filter(v => !isEmpty(v))
          .map(stringToSelectOpt),
    })
  }

  getClusters = () => {
    return this._request({
      method: "get",
      path: `/v1/clusters`,
      query: null,
      payload: null,
      preprocess: res =>
        get(res, "clusters", [])
          .filter(v => !isEmpty(v))
          .map(stringToSelectOpt),
    })
  }

  getTags = () => {
    return this._request({
      method: "get",
      path: `/v1/tags`,
      query: { limit: 5000 },
      payload: null,
      preprocess: res =>
        get(res, "tags", [])
          .filter(v => !isEmpty(v))
          .map(stringToSelectOpt),
    })
  }

  _constructURL = ({ path, query }) => {
    let q = ""

    if (!!query) {
      if (isString(query)) {
        q = query
      } else if (isObject(query)) {
        q = qs.stringify(query, { indices: false })
      }
    }

    return `${urljoin(this.location, path)}?${q}`
  }

  _request = ({ method, path, query, payload, preprocess }) => {
    return new Promise((resolve, reject) => {
      let config = { method, url: this._constructURL({ path, query }) }

      if (!isEmpty(payload)) {
        config.data = payload
      }

      axios(config)
        .then(res => {
          if (isFunction(preprocess)) {
            resolve(preprocess(res.data))
          }

          resolve(res.data)
        })
        .catch(error => {
          reject(error)
        })
    })
  }
}

export default FlotillaAPIClient
