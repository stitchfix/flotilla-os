import { createSlice } from "@reduxjs/toolkit"

type RunViewReducer = {
  shouldAutoscroll: boolean
  hasLogs: boolean
}

const initialState: RunViewReducer = {
  shouldAutoscroll: true,
  hasLogs: false,
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
  },
})

export const { toggleAutoscroll, setHasLogs } = runViewReducer.actions

export default runViewReducer.reducer
