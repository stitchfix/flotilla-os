import React from "react"
import PropTypes from "prop-types"
import styled from "styled-components"
import colors from "../../constants/colors"
import Loader from "./Loader"

const FIELD_HEIGHT_PX = 28
const FIELD_EL_MARGIN_LEFT_PX = 8

const FieldContainer = styled.div`
  width: 100%;
  display: flex;
  flex-flow: column nowrap;
  justify-content: flex-start;
  align-items: flex-start;
  margin-bottom: ${FIELD_HEIGHT_PX}px;
  position: relative;
`

export const FieldLabel = styled.div`
  font-size: 0.9rem;
  text-transform: uppercase;
  font-weight: 500;
  margin-bottom: 4px;
  margin-left: ${FIELD_EL_MARGIN_LEFT_PX}px;
`

export const FieldDescription = styled.div`
  font-size: 0.9rem;
  margin-left: ${FIELD_EL_MARGIN_LEFT_PX}px;
  margin-top: 4px;
  color: ${colors.gray[0]};
`

const FieldLoaderContainer = styled.div`
  position: absolute;
  right: ${FIELD_EL_MARGIN_LEFT_PX}px;
  top: calc(21px + (${FIELD_HEIGHT_PX}px - 18px) / 2);
`

const Field = ({ label, children, description, error, isLoading }) => (
  <FieldContainer>
    {!!label && <FieldLabel>{label}</FieldLabel>}
    {children}
    {!!error && <div className="pl-form-group-error">{error}</div>}
    {!!description && <FieldDescription>{description}</FieldDescription>}
    {!!isLoading && (
      <FieldLoaderContainer>
        <Loader mini />
      </FieldLoaderContainer>
    )}
  </FieldContainer>
)

Field.displayName = "Field"

Field.propTypes = {
  children: PropTypes.node.isRequired,
  description: PropTypes.string,
  error: PropTypes.any,
  isLoading: PropTypes.bool.isRequired,
  label: PropTypes.string,
}

Field.defaultProps = {
  error: false,
  isLoading: false,
}

export default Field
