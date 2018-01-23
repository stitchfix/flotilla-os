import { ActionTypes } from '../constants/'

export function renderModal({ modal }) {
  return ({
    type: ActionTypes.RENDER_MODAL,
    payload: { modal }
  })
}

export function unrenderModal() {
  return ({ type: ActionTypes.UNRENDER_MODAL })
}
