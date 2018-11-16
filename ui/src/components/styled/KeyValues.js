import React from "react"
import PropTypes from "prop-types"
import styled from "styled-components"
import Field from "./Field"
import { SPACING_PX } from "../../constants/styles"

const KeyValuesContainer = styled.div`
  padding: ${({ depth }) => {
    if (depth === 0) {
      return `${SPACING_PX * 2}px`
    }

    return `${SPACING_PX * 2}px 0`
  }};
`

const KeyValuesLabel = styled.h3`
  margin-bottom: ${SPACING_PX * 1.5}px;
`

const KeyValues = ({ depth, items, label }) => (
  <KeyValuesContainer depth={depth}>
    {!!label && <KeyValuesLabel>{label}</KeyValuesLabel>}
    {Object.keys(items).map(key => (
      <Field label={key} key={key}>
        {items[key]}
      </Field>
    ))}
  </KeyValuesContainer>
)

KeyValues.propTypes = {
  depth: PropTypes.number.isRequired,
  items: PropTypes.objectOf(PropTypes.node),
  label: PropTypes.node,
}

KeyValues.defaultProps = {
  depth: 0,
  items: {},
}

export default KeyValues
