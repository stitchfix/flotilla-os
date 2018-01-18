import React, { Component } from "react"
import PropTypes from "prop-types"
import { formValues } from "redux-form"
import { get } from "lodash"
import { ReduxFormGroupSelect, withStateFetch } from "platforma"
import config from "../config"
import { getImageTagsEndpoint } from "../utils/"

export class ImageTagSelect extends Component {
  static propTypes = {
    image: PropTypes.string,
    fetch: PropTypes.func,
    data: PropTypes.shape({
      tags: PropTypes.arrayOf(PropTypes.string),
    }),
    error: PropTypes.any,
  }
  componentDidMount() {
    if (!!this.props.image) {
      this.fetchTags(this.props.image)
    }
  }
  componentWillReceiveProps(nextProps) {
    if (this.props.image !== nextProps.image) {
      this.fetchTags(nextProps.image)
    }
  }
  fetchTags(image) {
    const url = getImageTagsEndpoint(config.IMAGE_TAGS_ENDPOINT, image)

    this.props.fetch(url)
  }
  render() {
    const { isLoading, data, error } = this.props
    const options = get(data, "tags", [])

    return (
      <ReduxFormGroupSelect
        name="image_tag"
        label="Image Tag"
        isRequired
        isLoading={isLoading}
        disabled={options.length === 0}
        options={options.map(t => ({
          label: t,
          value: t,
        }))}
      />
    )
  }
}

export default formValues("image")(withStateFetch(ImageTagSelect))
