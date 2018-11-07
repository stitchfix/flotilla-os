import actionTypes from "../constants/actionTypes"

const renderModal = modal => ({
  type: actionTypes.RENDER_MODAL,
  payload: { modal },
})

const unrenderModal = () => ({
  type: actionTypes.UNRENDER_MODAL,
})

export default {
  renderModal,
  unrenderModal,
}
