import styled from "styled-components"

const ButtonGroup = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: flex-start;
  align-items: center;
  & > * {
    margin: 0 4px;
  }
`

export default ButtonGroup
