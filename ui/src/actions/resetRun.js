import { ActionTypes } from '../constants/'
import { clearRunInterval } from './'

export default function resetRun() {
  return (dispatch) => {
    Promise.resolve()
      .then(() => { dispatch({ type: ActionTypes.RESET_RUN }) })
      .then(() => { clearRunInterval() })
  }
}
