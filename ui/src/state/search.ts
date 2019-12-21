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
    incrementCursor(state) {
      if (state.matches !== null) {
        if (
          state.cursor === null ||
          state.cursor === state.matches.length - 1
        ) {
          state.cursor = 0
        } else {
          state.cursor = state.cursor + 1
        }
      }
    },
    decrementCursor(state) {
      if (state.matches !== null) {
        if (state.cursor === null || state.cursor === 0) {
          state.cursor = state.matches.length - 1
        } else {
          state.cursor = state.cursor + 1
        }
      }
    },
  },
})

export const { setSearchInputVisibility } = searchReducer.actions

export default searchReducer.reducer
