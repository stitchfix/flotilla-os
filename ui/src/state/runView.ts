import { createSlice, PayloadAction } from "@reduxjs/toolkit"

type RunViewReducer = {
  shouldAutoscroll: boolean
  hasLogs: boolean
  isLogRequestIntervalActive: boolean
}

const initialState: RunViewReducer = {
  shouldAutoscroll: true,
  hasLogs: false,
  isLogRequestIntervalActive: false,
}

const runViewReducer = createSlice({
  name: "runViewReducer",
  initialState: initialState,
  reducers: {
    toggleAutoscroll(state) {
      state.shouldAutoscroll = !state.shouldAutoscroll
    },

    setHasLogs(state) {
      state.hasLogs = true
    },

    toggleIsLogRequestIntervalActive(
      state,
      { payload }: PayloadAction<boolean>
    ) {
      state.isLogRequestIntervalActive = payload
    },
  },
})

export const {
  toggleAutoscroll,
  setHasLogs,
  toggleIsLogRequestIntervalActive,
} = runViewReducer.actions

export default runViewReducer.reducer
