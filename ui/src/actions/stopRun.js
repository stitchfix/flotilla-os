import { getApiRoot } from '../constants/'

// Note: the callback function is used to force the serverTableConnect
// HOC to refetch the task history after the task is stopped.
export default function stopRun({ taskID, runID }, cb) {
  return (dispatch) => {
    const url = `${getApiRoot()}/task/${taskID}/history/${runID}`
    fetch(url, { method: 'DELETE' })
      .then(res => res.json())
      .then((res) => {
        if (!!cb && typeof cb === 'function') {
          cb(res)
        }
      })
  }
}
