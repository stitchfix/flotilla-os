export default function getImageTagsEndpoint(str, image) {
  return str.replace(/(\{image\})/, image)
}
