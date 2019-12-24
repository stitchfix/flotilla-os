import { createSlice, PayloadAction } from "@reduxjs/toolkit"

// [line number, char index]
type SearchMatches = [number, number][]

// index of search matches
type SearchCursor = number

type SearchReducer = {
  isSearchInputVisible: boolean
  matches: SearchMatches
  cursor: SearchCursor
}

const initialState: SearchReducer = {
  isSearchInputVisible: false,
  matches: [],
  cursor: 0,
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
      state.cursor = 0
    },

    clearSearchState(state) {
      state.matches = []
      state.cursor = 0
    },

    incrementCursor(state) {
      if (state.matches.length > 0) {
        if (state.cursor === state.matches.length - 1) {
          state.cursor = 0
        } else {
          state.cursor = state.cursor + 1
        }
      }
    },

    decrementCursor(state) {
      if (state.matches.length > 0) {
        if (state.cursor === 0) {
          state.cursor = state.matches.length - 1
        } else {
          state.cursor = state.cursor - 1
        }
      }
    },
  },
})

export const {
  setSearchInputVisibility,
  incrementCursor,
  decrementCursor,
  setMatches,
  clearSearchState,
} = searchReducer.actions

export default searchReducer.reducer
