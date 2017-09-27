import config from '../config'

export const imageEndpoint = config.IMAGE_ENDPOINT
export const imageTagsEndpoint = image => config.IMAGE_TAGS_ENDPOINT.replace(/(\{image\})/, image)
