import React from "react"
import PropTypes from "prop-types"
import styled, { css } from "styled-components"
import colors from "../../constants/colors"
import intentTypes from "../../constants/intentTypes"
import Loader from "./Loader"

export const buttonStyles = css`
  background: ${({ intent }) => {
    switch (intent) {
      case intentTypes.primary:
        return colors.blue[0]
      case intentTypes.error:
        return colors.red[0]
      case intentTypes.warning:
        return colors.yellow[0]
      case intentTypes.success:
        return colors.green[0]
      default:
        return colors.black[3]
    }
  }};
  border-radius: 2px;
  border: 1px solid ${colors.black[3]};
  box-shadow: none;
  color: ${colors.light_gray[2]};
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 500;
  height: 34px;
  letter-spacing: 0.02rem;
  padding: 0 12px;
  text-transform: uppercase;
  transition-duration: 200ms;
  white-space: nowrap;
`

const StyledButton = styled.button`
  ${buttonStyles};
`

const Button = ({ isLoading, intent, isDisabled, type, onClick, children }) => {
  return (
    <StyledButton
      isDisabled={isDisabled}
      type={type}
      intent={intent}
      onClick={onClick}
    >
      {isLoading ? (
        <Loader
          mini
          spinnerStyle={{
            borderColor: colors.gray[2],
            borderLeftColor: colors.light_gray[3],
          }}
        />
      ) : (
        children
      )}
    </StyledButton>
  )
}

Button.displayName = "Button"

Button.propTypes = {
  children: PropTypes.node,
  intent: PropTypes.oneOf(Object.values(intentTypes)),
  isDisabled: PropTypes.bool.isRequired,
  isLoading: PropTypes.bool.isRequired,
  onClick: PropTypes.func,
  type: PropTypes.string.isRequired,
}

Button.defaultProps = {
  isDisabled: false,
  isLoading: false,
  type: "type",
}

export default Button
