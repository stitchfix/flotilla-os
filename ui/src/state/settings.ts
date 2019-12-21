import { createSlice, PayloadAction } from "@reduxjs/toolkit"
import { get } from "lodash"
import ls from "../localstorage"
import { LOCAL_STORAGE_SETTINGS_KEY } from "../constants"
import { AppThunk } from "./store"

export type Settings = {
  USE_OPTIMIZED_LOG_RENDERER: boolean
  SHOULD_OVERRIDE_CMD_F_IN_RUN_VIEW: boolean
  DEFAULT_OWNER_ID: string
}

type SettingsReducer = {
  isLoading: boolean
  isSettingsDialogOpen: boolean
  settings: Settings
}

const initialState: SettingsReducer = {
  isLoading: false,
  isSettingsDialogOpen: false,
  settings: {
    USE_OPTIMIZED_LOG_RENDERER: true,
    SHOULD_OVERRIDE_CMD_F_IN_RUN_VIEW: true,
    DEFAULT_OWNER_ID: "",
  },
}

const merge = (initial: Settings, cached: any): Settings => ({
  USE_OPTIMIZED_LOG_RENDERER: get(
    cached,
    "USE_OPTIMIZED_LOG_RENDERER",
    initial.USE_OPTIMIZED_LOG_RENDERER
  ),
  SHOULD_OVERRIDE_CMD_F_IN_RUN_VIEW: get(
    cached,
    "SHOULD_OVERRIDE_CMD_F_IN_RUN_VIEW",
    initial.SHOULD_OVERRIDE_CMD_F_IN_RUN_VIEW
  ),
  DEFAULT_OWNER_ID: get(cached, "DEFAULT_OWNER_ID", initial.DEFAULT_OWNER_ID),
})

const settingsReducer = createSlice({
  name: "settingsReducer",
  initialState: initialState,
  reducers: {
    initStart() {},
    initSuccess(state, { payload }: PayloadAction<any>) {
      state.settings = merge(state.settings, payload)
    },
    initFailure() {},
    updateStart(state) {
      state.isLoading = true
    },
    updateSuccess(state, { payload }: PayloadAction<Settings>) {
      state.isLoading = false
      state.settings = merge(state.settings, payload)
      state.isSettingsDialogOpen = false
    },
    updateFailure(state) {
      state.isLoading = false
    },
    toggleDialogVisibilityChange(
      state,
      { payload }: PayloadAction<boolean | undefined>
    ) {
      state.isSettingsDialogOpen =
        payload === undefined ? !state.isSettingsDialogOpen : payload
    },
  },
})

export const {
  initStart,
  initSuccess,
  initFailure,
  updateStart,
  updateSuccess,
  updateFailure,
  toggleDialogVisibilityChange,
} = settingsReducer.actions

export const init = (): AppThunk => async dispatch => {
  try {
    dispatch(initStart())
    const cached = await ls.getItem<any>(LOCAL_STORAGE_SETTINGS_KEY)
    dispatch(initSuccess(cached))
  } catch (error) {
    console.error("Failed to initialize app settings from cache.")
    dispatch(initFailure())
  }
}

export const update = (s: Settings): AppThunk => async dispatch => {
  try {
    dispatch(updateStart())
    const cached = await ls.setItem<Settings>(LOCAL_STORAGE_SETTINGS_KEY, s)
    dispatch(updateSuccess(cached))
  } catch (error) {
    console.error("Failed to initialize app settings from cache.")
    dispatch(updateFailure())
  }
}

export default settingsReducer.reducer
