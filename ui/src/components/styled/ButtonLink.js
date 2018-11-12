import styled from "styled-components"
import { Link } from "react-router-dom"
import { buttonStyles } from "./Button"

const ButtonLink = styled(Link)`
  ${buttonStyles};
  display: flex;
  flex-flow: row nowrap;
  justify-content: center;
  align-items: center;
`

export default ButtonLink
