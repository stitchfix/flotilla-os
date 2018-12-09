import React, { PureComponent } from "react"
import styled from "styled-components"
import colors from "../../helpers/colors"
import Loader from "./Loader"
import SecondaryText from "./SecondaryText"

const FIELD_HEIGHT_PX = 28
const FIELD_EL_MARGIN_LEFT_PX = 8

export const FieldContainer = styled.div`
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

  ${({ isRequired }: { isRequired: boolean }) => {
    if (isRequired) {
      return `
        &:after {
          content: "*";
          color: ${colors.red[0]};
          margin-left: 2px;
        }
      `
    }
  }};
`

export const FieldDescription = styled(SecondaryText)`
  margin-top: 8px;
`

export const FieldError = styled(FieldDescription)`
  color: ${colors.red[0]};
`

const FieldLoaderContainer = styled.div`
  position: absolute;
  right: ${FIELD_EL_MARGIN_LEFT_PX}px;
  top: calc(21px + (${FIELD_HEIGHT_PX}px - 18px) / 2);
`

const FieldChildren = styled.div`
  color: ${colors.gray[3]};
  width: 100%;
`

interface IFieldProps {
  description?: string
  error?: any
  isLoading: boolean
  isRequired: boolean
  label?: string
}

class Field extends PureComponent<IFieldProps> {
  static defaultProps: Partial<IFieldProps> = {
    isLoading: false,
    isRequired: false,
  }

  render() {
    const {
      label,
      children,
      description,
      error,
      isLoading,
      isRequired,
    } = this.props

    return (
      <FieldContainer>
        {!!label && <FieldLabel isRequired={isRequired}>{label}</FieldLabel>}
        <FieldChildren>{children}</FieldChildren>
        {!!error && <FieldError>{error}</FieldError>}
        {!!description && <FieldDescription>{description}</FieldDescription>}
        {!!isLoading && (
          <FieldLoaderContainer>
            <Loader />
          </FieldLoaderContainer>
        )}
      </FieldContainer>
    )
  }
}

export default Field
