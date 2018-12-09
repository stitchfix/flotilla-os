import styled, { css } from "styled-components"
import colors from "../../helpers/colors"

// See this discussion as to why secondaryTextStyles is typed as `any`:
// https://github.com/reakit/reakit/pull/297#issuecomment-443535755
export const secondaryTextStyles: any = css`
  font-size: 0.9rem;
  color: ${colors.gray[1]};
`

const SecondaryText = styled.div`
  ${secondaryTextStyles};
`

export default SecondaryText
