import * as React from "react"
import { DebounceInput } from "react-debounce-input"
import { ButtonGroup, Button } from "@blueprintjs/core"

type Props = {
  onChange: (value: string) => void
  onFocus: () => void
  onBlur: () => void
  onIncrement: () => void
  onDecrement: () => void
  inputRef: React.Ref<HTMLInputElement>
  cursorIndex: number
  totalMatches: number
}

const LogVirtualizedSearch: React.FC<Props> = ({
  onChange,
  onFocus,
  onBlur,
  inputRef,
  onIncrement,
  onDecrement,
  cursorIndex,
  totalMatches,
}) => (
  <div className="flotilla-logs-virtualized-search-container">
    <DebounceInput
      onChange={evt => {
        onChange(evt.target.value)
      }}
      debounceTimeout={500}
      className="bp3-input flotilla-logs-virtualized-search-input"
      inputRef={inputRef}
      onFocus={onFocus}
      onBlur={onBlur}
      placeholder="Search..."
    />
    {totalMatches > 0 && (
      <div className="flotilla-logs-virtualized-search-info">
        {cursorIndex + 1}/{totalMatches}
      </div>
    )}
    <ButtonGroup>
      <Button
        icon="chevron-left"
        onClick={onDecrement}
        minimal
        disabled={totalMatches === 0}
      />
      <Button
        icon="chevron-right"
        onClick={onIncrement}
        minimal
        disabled={totalMatches === 0}
      />
    </ButtonGroup>
  </div>
)

export default LogVirtualizedSearch
