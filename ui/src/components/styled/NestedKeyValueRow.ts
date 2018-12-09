import styled from "styled-components"

const NestedKeyValueRow = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: space-between;
  align-items: flex-start;
  width: 100%;

  & > * {
    margin-right: 6px;
    &:last-child {
      margin-right: 0;
    }

    &:nth-child(3) {
      transform: translateY(8px);
    }
  }
`

export default NestedKeyValueRow
