import { get } from "lodash"

class FlotillaAPIError {
  constructor(error) {
    if (error.response) {
      this.data = get(error.response.data, "error")
      this.status = error.response.status
      this.headers = error.response.headers
    } else if (error.request) {
      this.data = error.request
      this.status = null
      this.headers = null
    } else {
      this.data = error.message
      this.status = null
      this.headers = null
    }
  }

  getError() {
    return {
      data: this.data,
      status: this.status,
      headers: this.status,
    }
  }
}

export default FlotillaAPIError
