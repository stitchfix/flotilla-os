import styled, { css } from "styled-components"
import { MONOSPACE_FONT_FAMILY } from "../../constants/styles"

export const monospaceStyles = css`
  font-family: ${MONOSPACE_FONT_FAMILY};
  font-size: 0.9rem;
  white-space: pre-line;
  word-break: break-all;
`

export const Pre = styled.pre`
  ${monospaceStyles};
`

export const Code = styled.code`
  ${monospaceStyles};
`
