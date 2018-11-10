import React from "react"
import PropTypes from "prop-types"
import { isEmpty } from "lodash"
import styled from "styled-components"
import Button from "../styled/Button"
import ButtonGroup from "../styled/ButtonGroup"

const POPUP_DISTANCE_PX = 40

const popupAnimation = styled.keyframes`
  from {
    transform: translateY(100%);
    opacity: 0;
  }
  to {
    transform: translateY(0%);
    opacity: 1;
  }
`

const StyledPopupContainer = styled.div`
  position: fixed;
  bottom: ${POPUP_DISTANCE_PX}px;
  right: ${POPUP_DISTANCE_PX}px;
  z-index: 8000;
  animation-duration: 0.2s;
  animation-name: ${popupAnimation};
  animation-timing-function: ease;
`

const StyledPopup = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: flex-start;
  align-items: stretch;
`

const StyledPopupTitle = styled.div``

const Popup = ({ title, body, actions }) => {
  return (
    <StyledPopupContainer>
      <StyledPopup>
        <StyledPopupTitle>{title}</StyledPopupTitle>
        <div>{body}</div>
        {!isEmpty(actions) && (
          <ButtonGroup>
            {actions.map((action, i) => (
              <Button {...action} key={i}>
                {action.text}
              </Button>
            ))}
          </ButtonGroup>
        )}
      </StyledPopup>
    </StyledPopupContainer>
  )
}

Popup.displayName = "Popup"

Popup.propTypes = {
  actions: PropTypes.arrayOf(
    PropTypes.shape({
      text: PropTypes.node,
    })
  ),
  body: PropTypes.node,
  title: PropTypes.node.isRequired,
}

export default Popup
