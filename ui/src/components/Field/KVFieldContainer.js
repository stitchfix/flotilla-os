import React from "react"
import PropTypes from "prop-types"
import {
  FieldContainer,
  FieldLabel,
  FieldDescription,
  FieldError,
} from "../styled/Field"

const KVFieldContainer = props => (
  <FieldContainer>
    {!!props.label && (
      <FieldLabel isRequired={props.isRequired}>{props.label}</FieldLabel>
    )}
    {!!props.description && (
      <span style={{ marginBottom: 8, marginTop: 0 }}>
        <FieldDescription>{props.description}</FieldDescription>
      </span>
    )}
    {!!props.error && <FieldError>{props.error}</FieldError>}
    {props.children}
  </FieldContainer>
)

KVFieldContainer.propTypes = {
  children: PropTypes.node.isRequired,
  description: PropTypes.string,
  error: PropTypes.any,
  isRequired: PropTypes.bool.isRequired,
  label: PropTypes.string,
}

export default KVFieldContainer
