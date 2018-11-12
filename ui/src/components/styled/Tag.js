import styled from "styled-components"
import SecondaryText from "./SecondaryText"
import colors from "../../constants/colors"
import { MONOSPACE_FONT_FAMILY } from "../../constants/styles"

const Tag = styled(SecondaryText)`
  background: ${colors.black[4]};
  padding: 6px 8px;
  border-radius: 4px;
  font-family: ${MONOSPACE_FONT_FAMILY};
`

export default Tag
