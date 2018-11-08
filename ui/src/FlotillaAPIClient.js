import axios from "axios"
import { isEmpty, isString, isObject } from "lodash"
import qs from "qs"
import urljoin from "url-join"

/**
 * API client to communicate with the Flotilla API
 */
class FlotillaAPIClient {
  constructor({ location }) {
    this.location = location
    this.getTasks = this.getTasks.bind(this)
    this.getTask = this.getTask.bind(this)
    this.getTaskHistory = this.getTaskHistory.bind(this)
    this.getActiveRuns = this.getActiveRuns.bind(this)
    this.createTask = this.createTask.bind(this)
    this.updateTask = this.updateTask.bind(this)
  }

  getTasks(query = { offset: 0, limit: 20 }) {
    return this._request({
      method: "get",
      path: "/task",
      query,
      payload: null,
    })
  }

  getTask({ definitionID }) {
    return this._request({
      method: "get",
      path: `/task/${definitionID}`,
      query: null,
      payload: null,
    })
  }

  getTaskHistory(
    { definitionID, query } = { query: { limit: 20, offset: 0 } }
  ) {
    return this._request({
      method: "get",
      path: `/task/${definitionID}/history`,
      query,
      payload: null,
    })
  }

  getActiveRuns(query = { offset: 0, limit: 20 }) {
    return this._request({
      method: "get",
      path: "/history",
      query,
      payload: null,
    })
  }

  createTask({ values }) {
    return this._request({
      method: "post",
      path: "/task",
      query: null,
      payload: values,
    })
  }

  updateTask({ definitionID, values }) {
    return this._request({
      method: "put",
      path: `/task/${definitionID}`,
      query: null,
      payload: values,
    })
  }

  deleteTask({ definitionID }) {
    return this._request({
      method: "delete",
      path: `/task/${definitionID}`,
      query: null,
      payload: null,
    })
  }

  stopRun({ definitionID, runID }) {
    return this._request({
      method: "delete",
      path: `/task/${definitionID}/history/${runID}`,
      query: null,
      payload: null,
    })
  }

  _constructURL({ path, query }) {
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

  _request({ method, path, query, payload }) {
    return new Promise((resolve, reject) => {
      let config = { method, url: this._constructURL({ path, query }) }

      if (!isEmpty(payload)) {
        config.data = payload
      }

      axios(config)
        .then(res => {
          resolve(res.data)
        })
        .catch(error => {
          reject(error)
        })
    })
  }
}

export default FlotillaAPIClient
