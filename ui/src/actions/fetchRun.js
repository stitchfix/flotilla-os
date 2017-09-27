import { has } from 'lodash'
import { ActionTypes, getApiRoot, runStatusTypes } from '../constants/'
import { calculateTaskDuration, checkStatus } from '../utils/'

// Set global variables.
var pollRunInfo

const requestRunLogs = () => ({ type: ActionTypes.REQUEST_RUN_LOGS })
const receiveRunLogs = ({ lastSeen, logs }) => ({
  type: ActionTypes.RECEIVE_RUN_LOGS,
  payload: { lastSeen, logs }
})
const receiveRunLogsError = error => ({
  type: ActionTypes.RECEIVE_RUN_LOGS_ERROR,
  payload: { error }
})

const requestRunInfo = () => ({ type: ActionTypes.REQUEST_RUN_INFO })
export const receiveRunInfo = info => ({
  type: ActionTypes.RECEIVE_RUN_INFO,
  payload: { info }
})
const receiveRunInfoError = error => ({
  type: ActionTypes.RECEIVE_RUN_INFO_ERROR,
  payload: { error }
})

export const clearRunInterval = () => {
  console.log('Shutting down run interval.')
  window.clearInterval(pollRunInfo)
}

export function fetchRunLogs({ runID, clearIntervalFn }) {
  return (dispatch, getState) => {
    dispatch(requestRunLogs())

    const lastSeen = getState().run.lastSeen
    let url = `${getApiRoot()}/${runID}/logs`

    if (!!lastSeen) {
      url += `?last_seen=${lastSeen}`
    }

    return fetch(url)
      .then(checkStatus)
      .then(res => res.json())
      .then((res) => {
        if (!(res.last_seen === lastSeen && res.log === '')) {
          dispatch(receiveRunLogs({
            logs: res.log,
            lastSeen: res.last_seen
          }))
        }

        if (!lastSeen || res.last_seen !== lastSeen) {
          if (has(res, 'last_seen')) {
            dispatch(fetchRunLogs({ runID, clearIntervalFn: clearRunInterval }))
          }
        }

        if (getState().run.info.status.toLowerCase() === runStatusTypes.stopped) {
          clearIntervalFn()
          dispatch({ type: ActionTypes.STOP_RUN_INTERVAL })
        }
      })
      .catch((e) => { dispatch(receiveRunLogsError(e)) })
  }
}

export function fetchRunInfo({ runID, clearIntervalFn }) {
  return (dispatch) => {
    dispatch(requestRunInfo())

    return fetch(`${getApiRoot()}/task/history/${runID}`)
      .then(checkStatus)
      .then(res => res.json())
      .then((res) => {
        // Send the run's info/metadata back for the component to
        // render.
        dispatch(receiveRunInfo(res))

        // Then send a GET request for the run's logs only if the
        // run is not queued.
        if (res.status !== runStatusTypes.QUEUED) {
          dispatch(fetchRunLogs({ runID, clearIntervalFn }))
        }
      })
      .catch((err) => {
        clearIntervalFn()
        dispatch(receiveRunInfoError(err))
      })
  }
}

export default function fetchRun({ runID }) {
  return (dispatch) => {
    // Call it once then begin polling.
    dispatch(fetchRunInfo({ runID, clearIntervalFn: clearRunInterval }))
    pollRunInfo = window.setInterval(() => {
      dispatch(fetchRunInfo({ runID, clearIntervalFn: clearRunInterval }))
    }, 5000)
    return pollRunInfo
  }
}
