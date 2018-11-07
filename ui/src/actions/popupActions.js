import actionTypes from "../constants/actionTypes"

const renderPopup = popup => ({
  type: actionTypes.RENDER_POPUP,
  payload: popup,
})

const unrenderPopup = () => ({
  type: actionTypes.UNRENDER_POPUP,
})

export default {
  renderPopup,
  unrenderPopup,
}
