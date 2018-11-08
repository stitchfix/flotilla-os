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
    this.getActiveRuns = this.getActiveRuns.bind(this)
  }

  getTasks(query = { offset: 0, limit: 20 }) {
    return this._request({
      method: "get",
      path: "/task",
      query,
      payload: null,
    })
  }

  getActiveRuns(query = { offset: 0, limit: 20 }) {
    const q = `status=RUNNING&status=PENDING&status=QUEUED&${qs.stringify(
      query
    )}`

    return this._request({
      method: "get",
      path: "/history",
      query: q,
      payload: null,
    })
  }

  createTask({ values }) {}
  updateTask({ definitionID, values }) {}

  _constructURL({ path, query }) {
    let q = ""

    if (!!query) {
      if (isString(query)) {
        q = query
      } else if (isObject(query)) {
        q = qs.stringify(query)
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
