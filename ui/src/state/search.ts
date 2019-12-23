import { createSlice, PayloadAction } from "@reduxjs/toolkit"

// [line number, char index]
type SearchMatches = [number, number][] | null

// index of search matches
type SearchCursor = number | null

type SearchReducer = {
  isSearchInputVisible: boolean
  matches: SearchMatches
  cursor: SearchCursor
}

const initialState: SearchReducer = {
  isSearchInputVisible: false,
  matches: null,
  cursor: null,
}

const searchReducer = createSlice({
  name: "searchReducer",
  initialState: initialState,
  reducers: {
    setSearchInputVisibility(
      state,
      { payload }: PayloadAction<{ isVisible: boolean }>
    ) {
      state.isSearchInputVisible = payload.isVisible
    },
    setMatches(state, { payload }: PayloadAction<{ matches: SearchMatches }>) {
      state.matches = payload.matches
      state.cursor = null
    },
    clear(state) {
      state.matches = null
      state.cursor = null
    },
    incrementCursor({ matches, cursor }) {
      if (matches !== null) {
        if (cursor === null || cursor === matches.length - 1) {
          cursor = 0
        } else {
          cursor = cursor + 1
        }
      }
    },
    decrementCursor({ matches, cursor }) {
      if (matches !== null) {
        if (cursor === null || cursor === 0) {
          cursor = matches.length - 1
        } else {
          cursor = cursor + 1
        }
      }
    },
  },
})

export const { setSearchInputVisibility } = searchReducer.actions

export default searchReducer.reducer
