import * as React from "react"
import { DebounceInput } from "react-debounce-input"
import { ButtonGroup, Button, Spinner, Classes } from "@blueprintjs/core"

type Props = {
  onChange: (value: string) => void
  onFocus: () => void
  onBlur: () => void
  onIncrement: () => void
  onDecrement: () => void
  inputRef: React.Ref<HTMLInputElement> | null
  cursorIndex: number
  totalMatches: number
  isSearchProcessing: boolean
  searchQuery: string
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
  isSearchProcessing,
  searchQuery,
}) => (
  <div className="flotilla-logs-virtualized-search-container">
    <input
      onChange={evt => {
        onChange(evt.target.value)
      }}
      className="bp3-input flotilla-logs-virtualized-search-input"
      ref={inputRef}
      onFocus={onFocus}
      onBlur={onBlur}
      placeholder="Search..."
      value={searchQuery}
    />
    {isSearchProcessing ? (
      <Spinner size={Spinner.SIZE_SMALL} />
    ) : (
      totalMatches > 0 && (
        <div className="flotilla-logs-virtualized-search-info">
          {cursorIndex + 1}/{totalMatches}
        </div>
      )
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
