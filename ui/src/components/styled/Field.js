import React from "react"
import PropTypes from "prop-types"
import styled from "styled-components"
import colors from "../../constants/colors"

const FIELD_EL_MARGIN_LEFT_PX = 8

const FieldContainer = styled.div`
  width: 100%;
  display: flex;
  flex-flow: column nowrap;
  justify-content: flex-start;
  align-items: flex-start;
  margin-bottom: 28px;
`

export const FieldLabel = styled.div`
  font-size: 0.9rem;
  text-transform: uppercase;
  font-weight: 500;
  margin-bottom: 4px;
  margin-left: ${FIELD_EL_MARGIN_LEFT_PX}px;
`

const FieldDescription = styled.div`
  font-size: 0.9rem;
  margin-left: ${FIELD_EL_MARGIN_LEFT_PX}px;
  margin-top: 4px;
  color: ${colors.gray[0]};
`

const Field = ({ label, children, description, error }) => (
  <FieldContainer>
    {!!label && <FieldLabel>{label}</FieldLabel>}
    {children}
    {!!error && <div className="pl-form-group-error">{error}</div>}
    {!!description && <FieldDescription>{description}</FieldDescription>}
  </FieldContainer>
)

Field.displayName = "Field"

Field.propTypes = {
  children: PropTypes.node.isRequired,
  description: PropTypes.string,
  error: PropTypes.any,
  label: PropTypes.string,
}

Field.defaultProps = {
  error: false,
}

export default Field
