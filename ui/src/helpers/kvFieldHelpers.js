import PropTypes from "prop-types"

export const SHARED_KV_FIELD_PROPS = {
  field: PropTypes.string.isRequired,
  isKeyRequired: PropTypes.bool.isRequired,
  isValueRequired: PropTypes.bool.isRequired,
  keyField: PropTypes.string.isRequired,
  label: PropTypes.node.isRequired,
  valueField: PropTypes.string.isRequired,
  values: PropTypes.arrayOf(PropTypes.objects).isRequired,
}

export const SHARED_KV_FIELD_DEFAULT_PROPS = {
  isKeyRequired: true,
  isValueRequired: false,
  keyField: "name",
  valueField: "value",
}
