import colors from "./colors"

export const NAVIGATION_HEIGHT_PX = 64
export const SPACING_PX = 12
export const VIEW_HEADER_HEIGHT_PX = 64
export const DEFAULT_FONT_COLOR = colors.gray[4]
export const SECONDARY_FONT_COLOR = colors.gray[0]
export const DEFAULT_BORDER = `1px solid ${colors.black[4]}`
export const MONOSPACE_FONT_FAMILY = `"Source Code Pro", "Courier New", Courier, monospace`
export const BREAKPOINTS_PX = {
  XL: 1400,
  L: 1200,
  M: 1000,
  S: 800,
}
export const DETAIL_VIEW_SIDEBAR_WIDTH_PX = 400
export const RUN_BAR_HEIGHT_PX = 64
export const LOADER_SIZE_PX = 20
export const Z_INDICES = {
  VIEW_HEADER: 2000,
  NAVIGATION: 5000,
  MODAL_CONTAINER: 6000,
  MODAL_OVERLAY: 6100,
  MODAL: 6200,
  POPUP: 7000,
}
