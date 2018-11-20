import React from "react"
import PropTypes from "prop-types"
import styled, { css } from "styled-components"
import colors from "../../constants/colors"
import intentTypes from "../../constants/intentTypes"
import Loader from "./Loader"
import intentToColor from "../../utils/intentToColor"

export const buttonStyles = css`
  background: ${({ intent }) => intentToColor(intent)};
  border-radius: 2px;
  border: 1px solid ${({ intent }) => intentToColor(intent)};
  box-shadow: none;
  color: ${colors.light_gray[3]};
  cursor: pointer;
  font-size: 0.85rem;
  font-weight: 600;
  height: 30px;
  letter-spacing: 0.04rem;
  padding: 0 12px;
  text-transform: uppercase;
  transition-duration: 100ms;
  white-space: nowrap;

  &:hover {
    background: ${colors.light_gray[3]};
    border-color: ${colors.light_gray[3]};
    color: ${({ intent }) => intentToColor(intent)};
  }

  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
    background: ${colors.gray[2]} !important;
    border-color: ${colors.gray[2]} !important;
    color: ${colors.black[0]} !important;
  }
`

const StyledButton = styled.button`
  ${buttonStyles};
`

const Button = ({ isLoading, intent, isDisabled, type, onClick, children }) => {
  return (
    <StyledButton
      disabled={isDisabled}
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
