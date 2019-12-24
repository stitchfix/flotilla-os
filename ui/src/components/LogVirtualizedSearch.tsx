import * as React from "react"
import { useSelector, useDispatch } from "react-redux"
import { get } from "lodash"
import Query from "./QueryParams"
import { DebounceInput } from "react-debounce-input"
import { LOG_SEARCH_QUERY_KEY } from "../constants"
import { ButtonGroup, Button } from "@blueprintjs/core"
import { RootState } from "../state/store"
import { decrementCursor, incrementCursor } from "../state/search"

const LogVirtualizedSearch: React.FC = props => {
  const dispatch = useDispatch()

  return (
    <Query>
      {({ query, setQuery }) => {
        return (
          <div
            className="flotilla-logs-virtualized-search-container"
            style={{ display: "flex" }}
          >
            <DebounceInput
              value={get(query, LOG_SEARCH_QUERY_KEY, "")}
              onChange={evt => {
                setQuery({ [LOG_SEARCH_QUERY_KEY]: evt.target.value })
              }}
              debounceTimeout={500}
              className="bp3-input"
            />
            <ButtonGroup>
              <Button
                icon="chevron-left"
                onClick={() => {
                  dispatch(decrementCursor())
                }}
              />
              <Button
                icon="chevron-right"
                onClick={() => {
                  dispatch(incrementCursor())
                }}
              />
            </ButtonGroup>
          </div>
        )
      }}
    </Query>
  )
}

export default LogVirtualizedSearch
