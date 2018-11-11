import React from "react"
import PropTypes from "prop-types"
import styled from "styled-components"

const FormContainer = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: center;
  align-items: flex-start;
  width: 100%;
`

const FormInner = styled.div`
  width: 600px;
  padding-top: 24px;
  & > * {
    margin-bottom: 36px;
  }
`

const Form = ({ children }) => (
  <FormContainer>
    <FormInner>{children}</FormInner>
  </FormContainer>
)

Form.displayName = "Form"

Form.propTypes = {
  children: PropTypes.node,
}

export default Form
