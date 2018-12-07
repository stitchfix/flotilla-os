import React, { Component } from "react"
import PropTypes from "prop-types"
import styled from "styled-components"
import JSONView from "react-json-view"
import Field from "./Field"
import { SPACING_PX } from "../../helpers/styles"
import jsonViewProps from "../../helpers/reactJsonViewProps"
import Button from "./Button"
import ButtonGroup from "./ButtonGroup"

const KeyValuesContainer = styled.div`
  padding: ${({ depth }) => {
    if (depth === 0) {
      return `${SPACING_PX * 2}px`
    }

    return `${SPACING_PX * 2}px 0`
  }};
`

const KeyValuesHeader = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  height: 100%;
  margin-bottom: ${SPACING_PX * 1.5}px;
`

class KeyValues extends Component {
  state = {
    displayRawData: false,
  }

  onDisplayRawDataButtonClick = () => {
    this.setState(prevState => ({ displayRawData: !prevState.displayRawData }))
  }

  render() {
    const { actions, depth, items, label, raw } = this.props
    const { displayRawData } = this.state

    let content

    if (!!displayRawData && !!raw) {
      content = <JSONView {...jsonViewProps} src={raw} />
    } else {
      content = Object.keys(items).map(key => (
        <Field label={key} key={key}>
          {items[key]}
        </Field>
      ))
    }

    return (
      <KeyValuesContainer depth={depth}>
        <KeyValuesHeader>
          <h3>{label}</h3>
          <ButtonGroup>
            <Button onClick={this.onDisplayRawDataButtonClick}>JSON</Button>
            {!!actions && actions}
          </ButtonGroup>
        </KeyValuesHeader>
        {content}
      </KeyValuesContainer>
    )
  }
}

KeyValues.propTypes = {
  actions: PropTypes.node,
  depth: PropTypes.number.isRequired,
  items: PropTypes.objectOf(PropTypes.node),
  label: PropTypes.node,
  raw: PropTypes.any,
}

KeyValues.defaultProps = {
  depth: 0,
  items: {},
}

export default KeyValues
