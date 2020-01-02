import * as React from "react"
import { useDispatch, useSelector } from "react-redux"
import { Switch } from "@blueprintjs/core"
import { RootState } from "../state/store"
import { toggleAutoscroll } from "../state/runView"

const AutoscrollSwitch: React.FC = () => {
  const dispatch = useDispatch()
  const shouldAutoscroll = useSelector(
    (state: RootState) => state.runView.shouldAutoscroll
  )

  return (
    <Switch
      checked={shouldAutoscroll}
      onChange={() => {
        dispatch(toggleAutoscroll())
      }}
    />
  )
}

export default AutoscrollSwitch
