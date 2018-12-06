import styled, { css } from "styled-components"
import colors from "../../helpers/colors"
import {
  DEFAULT_FONT_COLOR,
  SECONDARY_FONT_COLOR,
} from "../../helpers/styles"
import { monospaceStyles } from "./Monospace"

const sharedInputStyles = css`
  border-color: ${colors.blue[0]};
  background: ${colors.black[1]};
  border: 2px solid ${colors.black[3]};
  border-radius: 4px;
  font-size: 1rem;
  width: 100%;
  color: ${DEFAULT_FONT_COLOR};
  height: 40px;
  padding: 0 8px;
  &::placeholder {
    color: ${SECONDARY_FONT_COLOR};
  }
  &:focus {
    border-color: ${colors.blue[0]};
  }
`

export const Input = styled.input`
  ${sharedInputStyles};
`
export const Textarea = styled.textarea`
  ${sharedInputStyles} ${monospaceStyles} height: 240px;
`
