import axios from "axios"
import { isEmpty } from "lodash"
import qs from "qs"

/**
 * API client to communicate with the Flotilla API
 */
class FlotillaAPIClient {
  constructor({ location }) {
    this.location = location
    this.getTasks = this.getTasks.bind(this)
  }

  getTasks(query = { offset: 0, limit: 20 }) {
    return this._request({
      method: "get",
      path: "/task",
      query,
      payload: null,
    })
  }

  _constructURL({ path, query }) {
    return `${this.location}/${path}?${qs.stringify(query)}`
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
