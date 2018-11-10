import axios from "axios"
import { has } from "lodash"
import actionTypes from "../constants/actionTypes"
import config from "../config"

const mapStringArrayToSelectObjectArray = (obj, key) => {
  if (has(obj, key) && Array.isArray(obj[key])) {
    return obj[key].map(opt => ({ label: opt, value: opt }))
  }
  return []
}

const requestDropdownOpts = () => ({ type: actionTypes.REQUEST_SELECT_OPTS })
const receiveDropdownOpts = res => {
  return {
    type: actionTypes.RECEIVE_SELECT_OPTS,
    payload: {
      group: mapStringArrayToSelectObjectArray(res[0], "groups"),
      cluster: mapStringArrayToSelectObjectArray(res[1], "clusters"),
      tag: mapStringArrayToSelectObjectArray(res[2], "tags"),
    },
  }
}
const receiveDropdownOptsError = error => ({
  type: actionTypes.RECEIVE_SELECT_OPTS_ERROR,
  payload: error,
  error: true,
})

const fetchDropdownOpts = () => dispatch => {
  dispatch(requestDropdownOpts())

  axios
    .all([
      axios.get(`${config.FLOTILLA_API}/v1/groups?limit=2000`),
      axios.get(`${config.FLOTILLA_API}/v1/clusters`),
      axios.get(`${config.FLOTILLA_API}/v1/tags?limit=5000`),
    ])
    .then(
      axios.spread((group, cluster, tag) => {
        dispatch(receiveDropdownOpts([group.data, cluster.data, tag.data]))
      })
    )
    .catch(err => {
      dispatch(receiveDropdownOptsError(err))
    })
}

export default fetchDropdownOpts
