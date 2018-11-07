import hsluv from "hsluv"

const generateLightnessVariant = (hsluvArr, level, interval = 1.75) => [
  hsluvArr[0],
  hsluvArr[1],
  hsluvArr[2] + interval * level,
]
const generateLightnessVariants = (name, hsluvArr, lightnessInterval) => ({
  [`${name}_0`]: hsluv.hsluvToHex(hsluvArr),
  [`${name}_1`]: hsluv.hsluvToHex(
    generateLightnessVariant(hsluvArr, 1, lightnessInterval)
  ),
  [`${name}_2`]: hsluv.hsluvToHex(
    generateLightnessVariant(hsluvArr, 2, lightnessInterval)
  ),
  [`${name}_3`]: hsluv.hsluvToHex(
    generateLightnessVariant(hsluvArr, 3, lightnessInterval)
  ),
  [`${name}_4`]: hsluv.hsluvToHex(
    generateLightnessVariant(hsluvArr, 4, lightnessInterval)
  ),
})

export default {
  black: generateLightnessVariants("black", hsluv.hexToHsluv("#23292e"), 2.5),
  gray: generateLightnessVariants("gray", hsluv.hexToHsluv("#626f7a"), 10),
  light_gray: generateLightnessVariants(
    "light_gray",
    hsluv.hexToHsluv("#d0d9e1"),
    3
  ),
  blue: generateLightnessVariants("blue", hsluv.hexToHsluv("#58a7f3"), 3),
  green: generateLightnessVariants("green", hsluv.hexToHsluv("#4acccf"), 3),
  red: generateLightnessVariants("red", hsluv.hexToHsluv("#d76262"), 4.5),
  yellow: generateLightnessVariants("yellow", hsluv.hexToHsluv("#e3ca00"), 3),
}
