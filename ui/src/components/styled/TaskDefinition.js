import s from "styled-components"

export const TaskDefinitionView = s.div`
  display: flex;
  flex-flow: column nowrap;
  justify-content: flex-start;
  align-items: flex-start;
  width: 100%;
`

export const TaskDefinitionViewSidebar = s.div`
  position: fixed;
  top: 0;
  bottom: 0;
  right: 0;
`
