import styled, { css } from "styled-components"
import colors from "../../helpers/colors"

export const secondaryTextStyles = css`
  font-size: 0.9rem;
  color: ${colors.gray[1]};
`

const SecondaryText = styled.div`
  ${secondaryTextStyles};
`

export default SecondaryText
