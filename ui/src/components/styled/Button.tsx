import * as React from "react"
import styled, { css } from "styled-components"
import colors from "../../helpers/colors"
import Loader from "./Loader"
import intentToColor from "../../helpers/intentToColor"
import { intents } from "../../.."

export const buttonStyles = css`
  background: ${({ intent }: { intent?: intents }) => intentToColor(intent)};
  border-radius: 2px;
  border: 1px solid
    ${({ intent }: { intent?: intents }) => intentToColor(intent)};
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
    color: ${({ intent }: { intent?: intents }) => intentToColor(intent)};
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

interface IButtonProps {
  children: React.ReactNode
  intent?: intents
  isDisabled: boolean
  isLoading: boolean
  onClick?: (evt: React.SyntheticEvent) => void
  type: string
}

class Button extends React.PureComponent<IButtonProps> {
  static displayName = "Button"
  static defaultProps: Partial<IButtonProps> = {
    isDisabled: false,
    isLoading: false,
    type: "button",
  }
  render() {
    const {
      isLoading,
      intent,
      isDisabled,
      type,
      onClick,
      children,
    } = this.props
    return (
      <StyledButton
        disabled={isDisabled}
        type={type}
        intent={intent}
        onClick={onClick}
      >
        {isLoading ? <Loader /> : children}
      </StyledButton>
    )
  }
}

export default Button
