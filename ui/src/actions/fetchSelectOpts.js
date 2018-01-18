import React from "react"
import { intentTypes, Popup, popupActions } from "aa-ui-components"
import axios from "axios"
import config from "../config"
import { actionTypes } from "../constants/"

const strToSelectOpt = opt => ({ label: opt, value: opt })

const requestDropdownOpts = () => ({ type: actionTypes.REQUEST_SELECT_OPTS })

const receiveDropdownOpts = res => ({
  type: actionTypes.RECEIVE_SELECT_OPTS,
  payload: {
    image: res[0].repositories.map(strToSelectOpt),
    group: res[1].groups.map(strToSelectOpt),
    cluster: res[2].registered_clusters.map(strToSelectOpt),
    tag: res[3].tags.map(strToSelectOpt),
  },
})

const receiveDropdownOptsError = error => dispatch => {
  dispatch(
    popupActions.renderPopup(
      <Popup
        title="Error fetching select options!"
        message={error.toString()}
        intent={intentTypes.error}
        hide={() => {
          dispatch(popupActions.unrenderPopup())
        }}
        autohide={false}
      />
    )
  )
}

export default function fetchDropdownOpts() {
  return dispatch => {
    dispatch(requestDropdownOpts())

    axios
      .all([
        axios.get(config.IMAGE_ENDPOINT),
        axios.get(`${config.FLOTILLA_API}/groups?limit=2000`),
        axios.get(config.CLUSTERS_API),
        axios.get(`${config.FLOTILLA_API}/tags?limit=5000`),
      ])
      .then(
        axios.spread((image, group, cluster, tag) => {
          dispatch(
            receiveDropdownOpts([
              image.data,
              group.data,
              cluster.data,
              tag.data,
            ])
          )
        })
      )
      .catch(err => {
        dispatch(receiveDropdownOptsError(err))
      })
  }
}
