import { ActionTypes } from '../constants/'

export default function resetTask() {
  return ({ type: ActionTypes.RESET_TASK })
}
