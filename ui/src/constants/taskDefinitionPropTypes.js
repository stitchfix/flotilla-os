import PropTypes from "prop-types"

export default {
  env: PropTypes.arrayOf(
    PropTypes.shape({
      name: PropTypes.string,
      value: PropTypes.string,
    })
  ),
  arn: PropTypes.string,
  definition_id: PropTypes.string,
  image: PropTypes.string,
  group_name: PropTypes.string,
  container_name: PropTypes.string,
  user: PropTypes.string,
  alias: PropTypes.string,
  memory: PropTypes.number,
  command: PropTypes.string,
  tags: PropTypes.arrayOf(PropTypes.string),
}
