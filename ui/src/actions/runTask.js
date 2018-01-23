import { push } from 'react-router-redux'
import { getApiRoot } from '../constants/'

export default function runTask({ cluster, env, taskID }) {
  return (dispatch) => {
    const url = `${getApiRoot()}/task/${taskID}/execute`
    const body = { cluster }

    if (!!env) { body.env = env }

    fetch(url, {
      method: 'PUT',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify(body)
    })
      .then(res => res.json())
      .then((res) => {
        dispatch(push(`/runs/${res.run_id}`))
      })
  }
}
