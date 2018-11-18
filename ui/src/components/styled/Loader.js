import React from "react"
import styled, { keyframes } from "styled-components"
import colors from "../../constants/colors"
import { LOADER_SIZE_PX } from "../../constants/styles"

const LOADER_BORDER_WIDTH_PX = LOADER_SIZE_PX / 6
const LOADER_BORDER = `${LOADER_BORDER_WIDTH_PX}px solid ${colors.black[3]}`

const loadingAnimation = keyframes`
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
`

const LoaderContainer = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: center;
  align-items: center;
  width: 100%;
`

const LoaderInner = styled.div`
  animation: ${loadingAnimation} 1.4s infinite linear;
  border-bottom: ${LOADER_BORDER};
  border-radius: 50%;
  border-right: ${LOADER_BORDER};
  border-top: ${LOADER_BORDER};
  border-left: ${LOADER_BORDER_WIDTH_PX}px solid ${colors.blue[0]};
  height: ${LOADER_SIZE_PX}px;
  position: relative;
  text-indent: -9999em;
  transform: translateZ(0);
  width: ${LOADER_SIZE_PX}px;

  &:after {
    border-radius: 50%;
    width: ${LOADER_SIZE_PX}px;
    height: ${LOADER_SIZE_PX}px;
  }
`

const Loader = () => (
  <LoaderContainer>
    <LoaderInner />
  </LoaderContainer>
)

export default Loader
