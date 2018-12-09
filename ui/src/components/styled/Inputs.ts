import styled from "styled-components"
import colors from "../../helpers/colors"
import {
  DEFAULT_FONT_COLOR,
  SECONDARY_FONT_COLOR,
  MONOSPACE_FONT_FAMILY,
} from "../../helpers/styles"

export const Input = styled.input`
  background: ${colors.black[1]};
  border-color: ${colors.blue[0]};
  border-radius: 4px;
  border: 2px solid ${colors.black[3]};
  color: ${DEFAULT_FONT_COLOR};
  font-size: 1rem;
  height: 40px;
  padding: 0 8px;
  width: 100%;
  &::placeholder {
    color: ${SECONDARY_FONT_COLOR};
  }
  &:focus {
    border-color: ${colors.blue[0]};
  }
`
export const Textarea = styled.textarea`
  background: ${colors.black[1]};
  border-color: ${colors.blue[0]};
  border-radius: 4px;
  border: 2px solid ${colors.black[3]};
  color: ${DEFAULT_FONT_COLOR};
  font-family: ${MONOSPACE_FONT_FAMILY};
  font-size: 0.9rem;
  height: 240px;
  padding: 0 8px;
  white-space: pre-line;
  width: 100%;
  word-break: break-all;
  &::placeholder {
    color: ${SECONDARY_FONT_COLOR};
  }
  &:focus {
    border-color: ${colors.blue[0]};
  }
`
