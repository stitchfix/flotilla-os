import React, { Component } from "react"
import PropTypes from "prop-types"
import styled from "styled-components"
import PopupContext from "./PopupContext"
import { Z_INDICES, SPACING_PX, DEFAULT_BORDER } from "../../constants/styles"
import colors from "../../constants/colors"
import intentTypes from "../../constants/intentTypes"
import Card from "../styled/Card"
import Button from "../styled/Button"
import ButtonGroup from "../styled/ButtonGroup"

const POPUP_WINDOW_DISTANCE_PX = 48
const POPUP_WIDTH_PX = 400

const PopupPositioner = styled.div`
  position: fixed;
  bottom: ${POPUP_WINDOW_DISTANCE_PX}px;
  right: ${POPUP_WINDOW_DISTANCE_PX}px;
  z-index: ${Z_INDICES.POPUP};
  width: ${POPUP_WIDTH_PX}px;
`

class Popup extends Component {
  componentDidMount() {
    const { shouldAutohide, unrenderPopup, visibleDuration } = this.props

    if (!!shouldAutohide) {
      window.setTimeout(() => {
        unrenderPopup()
      }, visibleDuration)
    }
  }

  renderActions = () => {
    const { actions, unrenderPopup } = this.props

    return (
      <ButtonGroup>
        <Button onClick={unrenderPopup}>Close</Button>
        {!!actions && actions}
      </ButtonGroup>
    )
  }

  render() {
    const { body, title } = this.props

    return (
      <PopupPositioner>
        <Card title={title} actions={this.renderActions()}>
          {body}
        </Card>
      </PopupPositioner>
    )
  }
}

Popup.propTypes = {
  actions: PropTypes.node,
  body: PropTypes.node,
  intent: PropTypes.oneOf(Object.values(intentTypes)),
  shouldAutohide: PropTypes.bool.isRequired,
  title: PropTypes.node,
  unrenderPopup: PropTypes.func.isRequired,
  visibleDuration: PropTypes.number.isRequired,
}

Popup.defaultProps = {
  shouldAutohide: true,
  visibleDuration: 5000,
}

export default props => (
  <PopupContext.Consumer>
    {pCtx => <Popup {...props} unrenderPopup={pCtx.unrenderPopup} />}
  </PopupContext.Consumer>
)
