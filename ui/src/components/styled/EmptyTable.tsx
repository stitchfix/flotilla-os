import * as React from "react"
import styled from "styled-components"
import Loader from "./Loader"
import { SPACING_PX } from "../../helpers/styles"

const EmptyTableContainer = styled.div`
  display: flex;
  flex-flow: column nowrap;
  justify-content: center;
  align-items: center;
  width: 100%;
  height: 100%;
  padding: ${SPACING_PX * 2}px;
`

const EmptyTableTitle = styled.h2`
  margin-bottom: ${SPACING_PX * 2}px;
`

interface IEmptyTableProps {
  actions?: React.ReactNode
  error: boolean
  isLoading: boolean
  title?: React.ReactNode
}

class EmptyTable extends React.PureComponent<IEmptyTableProps> {
  static displayName = "EmptyTable"
  static defaultProps = {
    isLoading: false,
    error: false,
  }
  render() {
    const { isLoading, title, actions } = this.props
    let content

    if (isLoading) {
      content = <Loader />
    } else {
      content = (
        <React.Fragment>
          {title && <EmptyTableTitle>{title}</EmptyTableTitle>}
          {actions && <h2>{actions}</h2>}
        </React.Fragment>
      )
    }

    return <EmptyTableContainer>{content}</EmptyTableContainer>
  }
}

export default EmptyTable
