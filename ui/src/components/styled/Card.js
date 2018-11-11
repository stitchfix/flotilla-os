import React from "react"
import PropTypes from "prop-types"
import styled from "styled-components"
import colors from "../../constants/colors"
import { DEFAULT_BORDER, SPACING_PX } from "../../constants/styles"
import ButtonGroup from "./ButtonGroup"

const CARD_HEADER_FOOTER_HEIGHT_PX = 48

const CardContainer = styled.div`
  background: ${colors.black[0]};
  border: ${DEFAULT_BORDER};
  width: 100%;
  border-radius: 8px;
`

const CardHeader = styled.div`
  border-bottom: ${DEFAULT_BORDER};
  background: inherit;
  min-height: ${CARD_HEADER_FOOTER_HEIGHT_PX}px;
  width: 100%;
  padding: ${SPACING_PX}px;
  display: flex;
  flex-flow: row nowrap;
  justify-content: space-between;
  align-items: center;
`

const CardContent = styled.div`
  padding: ${SPACING_PX}px;
`

const CardFooter = styled.div`
  border-top: ${DEFAULT_BORDER};
  background: inherit;
  min-height: ${CARD_HEADER_FOOTER_HEIGHT_PX}px;
  width: 100%;
  padding: ${SPACING_PX}px;
`

const Card = ({ title, actions, footerActions, children }) => {
  const shouldRenderHeader = !!title || !!actions
  const shouldRenderFooter = !!footerActions

  return (
    <CardContainer>
      {shouldRenderHeader && (
        <CardHeader>
          <div>{title}</div>
          <ButtonGroup>{actions}</ButtonGroup>
        </CardHeader>
      )}
      <CardContent>{children}</CardContent>
      {shouldRenderFooter && (
        <CardFooter>
          <ButtonGroup>{footerActions}</ButtonGroup>
        </CardFooter>
      )}
    </CardContainer>
  )
}

Card.displayName = "Card"
Card.propTypes = {
  actions: PropTypes.node,
  children: PropTypes.node,
  footerActions: PropTypes.node,
  title: PropTypes.node,
}

export default Card
