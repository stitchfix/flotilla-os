import React, { ReactNode, PureComponent } from "react"
import styled from "styled-components"
import HeaderText from "./HeaderText"
import { SPACING_PX } from "../../helpers/styles"

const FormContainer = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: center;
  align-items: flex-start;
  width: 100%;
`

const FormInner = styled.div`
  display: flex;
  flex-flow: column nowrap;
  justify-content: flex-start;
  align-items: flex-start;
  width: 600px;
  padding-top: 24px;
  padding-bottom: ${SPACING_PX * 20}px;
  & > * {
    margin-bottom: 36px;
  }
`
interface IFormProps {
  title?: ReactNode
}

class Form extends PureComponent<IFormProps> {
  render() {
    const { children, title } = this.props

    return (
      <FormContainer>
        <FormInner>
          {!!title && <HeaderText>{title}</HeaderText>}
          {children}
        </FormInner>
      </FormContainer>
    )
  }
}

export default Form
