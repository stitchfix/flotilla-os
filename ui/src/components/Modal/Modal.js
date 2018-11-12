import React from "react"
import styled from "styled-components"
import { Z_INDICES } from "../../constants/styles"

const Modal = styled.div`
  z-index: ${Z_INDICES.MODAL};
  width: 400px;
  margin-top: 24px;
`

export default Modal
