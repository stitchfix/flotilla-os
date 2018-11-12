import styled from "styled-components"

const TagGroup = styled.div`
  display: flex;
  flex-flow: row wrap;
  justify-content: flex-start;
  align-items: flex-start;
  padding-left: 6px;
  padding-bottom: 6px;
  & > * {
    margin-top: 6px;
    margin-right: 6px;
  }
`

export default TagGroup
