import React from "react"
import PropTypes from "prop-types"
import styled from "styled-components"
import HeaderText from "./HeaderText"
import { SPACING_PX } from "../../constants/styles"

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

const Form = ({ children, title }) => (
  <FormContainer>
    <FormInner>
      {!!title && <HeaderText>{title}</HeaderText>}
      {children}
    </FormInner>
  </FormContainer>
)

Form.displayName = "Form"

Form.propTypes = {
  children: PropTypes.node,
  title: PropTypes.string,
}

Form.defaultProps = {
  title: "",
}

export default Form
