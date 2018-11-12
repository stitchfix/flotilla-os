import colors from "./colors"

export const TOPBAR_HEIGHT_PX = 64
export const SPACING_PX = 12
export const VIEW_HEADER_HEIGHT_PX = 64
export const DEFAULT_FONT_COLOR = colors.gray[4]
export const SECONDARY_FONT_COLOR = colors.gray[0]
export const DEFAULT_BORDER = `1px solid ${colors.black[2]}`
export const MONOSPACE_FONT_FAMILY = `"Source Code Pro", "Courier New", Courier, monospace`

export const Z_INDICES = {
  VIEW_HEADER: 2000,
  NAVIGATION: 5000,
  MODAL_CONTAINER: 6000,
  MODAL_OVERLAY: 6100,
  MODAL: 6200,
  POPUP: 7000,
}
