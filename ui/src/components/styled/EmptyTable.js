import React, { Fragment } from "react"
import PropTypes from "prop-types"
import styled from "styled-components"
import { has } from "lodash"
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

const EmptyTable = props => {
  let content

  if (props.isLoading) {
    content = <Loader />
  } else {
    content = (
      <Fragment>
        {props.title && <EmptyTableTitle>{props.title}</EmptyTableTitle>}
        {props.actions && <h2>{props.actions}</h2>}
      </Fragment>
    )
  }

  return <EmptyTableContainer>{content}</EmptyTableContainer>
}

EmptyTable.propTypes = {
  actions: PropTypes.node,
  error: PropTypes.bool.isRequired,
  isLoading: PropTypes.bool.isRequired,
  title: PropTypes.node,
}

EmptyTable.defaultProps = {
  isLoading: false,
  error: false,
}

EmptyTable.displayName = "EmptyTable"

export default EmptyTable
